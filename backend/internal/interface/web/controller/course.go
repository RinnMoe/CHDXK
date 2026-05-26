package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

type CourseController struct {
	query   *application.CourseQueryService
	command *application.CourseCommandService
}

func NewCourseController(query *application.CourseQueryService, command *application.CourseCommandService) *CourseController {
	return &CourseController{query: query, command: command}
}

func (ctrl *CourseController) GetCourseFilters(c *gin.Context) {
	result, err := ctrl.query.GetCourseFilters(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListCourses(c *gin.Context) {
	var f application.CourseListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.query.ListCourses(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListHotCourses(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		respondBadRequest(c, "数量限制无效")
		return
	}
	result, err := ctrl.query.ListHotCourses(c.Request.Context(), period, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) GetCourse(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		respondBadRequest(c, "课程 ID 无效")
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())

	result, err := ctrl.query.GetCourseDetail(c.Request.Context(), u, courseID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) SetNotificationLevel(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		respondBadRequest(c, "课程 ID 无效")
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var req struct {
		Level int `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err)
		return
	}
	if req.Level < 0 || req.Level > 2 {
		respondBadRequest(c, "通知级别必须是 0、1 或 2")
		return
	}

	if err := ctrl.command.SetNotificationLevel(c.Request.Context(), u.ID, courseID, course.NotificationLevel(req.Level)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *CourseController) ListFollowedCourses(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var f application.CourseListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.query.ListCoursesByNotificationLevel(c.Request.Context(), u.ID, course.NotificationLevelFollow, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListIgnoredCourses(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var f application.CourseListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.query.ListCoursesByNotificationLevel(c.Request.Context(), u.ID, course.NotificationLevelIgnored, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
