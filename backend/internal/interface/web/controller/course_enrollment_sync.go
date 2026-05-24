package controller

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

const enrollmentSyncSessionKey = "course_enrollment_sync_state"

type CourseEnrollmentSyncController struct {
	service             *application.CourseEnrollmentSyncService
	frontendCallbackURL string
}

func NewCourseEnrollmentSyncController(service *application.CourseEnrollmentSyncService, conf config.JAccountConfig) *CourseEnrollmentSyncController {
	return &CourseEnrollmentSyncController{
		service:             service,
		frontendCallbackURL: strings.TrimSpace(conf.FrontendCallbackURL),
	}
}

func (ctrl *CourseEnrollmentSyncController) Start(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	semester := strings.TrimSpace(c.Query("semester"))
	if semester == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "semester required"})
		return
	}
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	semester, authURL, err := ctrl.service.Start(c.Request.Context(), semester, state)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrSemesterRequired), errors.Is(err, application.ErrInvalidSyncSemester):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, application.ErrEnrollmentSyncDisabled):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	stored := enrollmentSyncState{State: state, UserID: u.ID, Semester: semester, CreatedAt: time.Now().Unix()}
	encoded, err := json.Marshal(stored)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s := sessions.Default(c)
	s.Set(enrollmentSyncSessionKey, string(encoded))
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (ctrl *CourseEnrollmentSyncController) Callback(c *gin.Context) {
	stored, err := loadEnrollmentSyncState(c)
	if err != nil {
		ctrl.redirectResult(c, "", "error", "会话已失效，请重新同步", 0, 0)
		return
	}
	if c.Query("state") == "" || c.Query("state") != stored.State {
		ctrl.redirectResult(c, stored.Semester, "error", "登录状态校验失败，请重新同步", 0, 0)
		return
	}
	if oauthError := strings.TrimSpace(c.Query("error")); oauthError != "" {
		ctrl.redirectResult(c, stored.Semester, "error", oauthError, 0, 0)
		return
	}
	code := strings.TrimSpace(c.Query("code"))
	if code == "" {
		ctrl.redirectResult(c, stored.Semester, "error", "缺少授权 code", 0, 0)
		return
	}

	result, err := ctrl.service.SyncFromCode(c.Request.Context(), stored.UserID, stored.Semester, code)
	if err != nil {
		ctrl.redirectResult(c, stored.Semester, "error", "同步选课记录失败", 0, 0)
		return
	}
	clearEnrollmentSyncState(c)
	ctrl.redirectResult(c, stored.Semester, "ok", "", result.Matched, result.Total)
}

func (ctrl *CourseEnrollmentSyncController) redirectResult(c *gin.Context, semester, status, message string, matched int64, total int) {
	callback := ctrl.frontendCallbackURL
	if callback == "" {
		callback = "/course/mine/sync-callback"
	}
	u, err := url.Parse(callback)
	if err != nil {
		c.String(http.StatusInternalServerError, "invalid frontend callback url")
		return
	}
	q := u.Query()
	q.Set("status", status)
	if semester != "" {
		q.Set("semester", semester)
	}
	if message != "" {
		q.Set("message", message)
	}
	if matched > 0 {
		q.Set("matched", strconv.FormatInt(matched, 10))
	}
	if total > 0 {
		q.Set("total", strconv.Itoa(total))
	}
	u.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, u.String())
}

type enrollmentSyncState struct {
	State     string `json:"state"`
	UserID    int    `json:"user_id"`
	Semester  string `json:"semester"`
	CreatedAt int64  `json:"created_at"`
}

func loadEnrollmentSyncState(c *gin.Context) (*enrollmentSyncState, error) {
	v, ok := sessions.Default(c).Get(enrollmentSyncSessionKey).(string)
	if !ok || v == "" {
		return nil, fmt.Errorf("missing enrollment sync state")
	}
	var state enrollmentSyncState
	if err := json.Unmarshal([]byte(v), &state); err != nil {
		return nil, err
	}
	if state.State == "" || state.UserID == 0 || state.Semester == "" {
		return nil, fmt.Errorf("invalid enrollment sync state")
	}
	return &state, nil
}

func clearEnrollmentSyncState(c *gin.Context) {
	s := sessions.Default(c)
	s.Delete(enrollmentSyncSessionKey)
	_ = s.Save()
}

func randomState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
