package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

type AdminUserController struct {
	query   *application.AdminUserQueryService
	command *application.AdminUserCommandService
}

type suspendUserCommand struct {
	Days int `json:"days"`
}

type resetPasswordCommand struct {
	Password string `json:"password" binding:"required"`
}

type adminUserQuery struct {
	Email    string `form:"email"`
	Username string `form:"username"`
	ReviewID int    `form:"review_id"`
}

func NewAdminUserController(
	query *application.AdminUserQueryService,
	command *application.AdminUserCommandService,
) *AdminUserController {
	return &AdminUserController{query: query, command: command}
}

func (ctrl *AdminUserController) GetUser(c *gin.Context) {
	var query adminUserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondBindError(c, err)
		return
	}
	if query.ReviewID < 0 {
		respondBadRequest(c, "点评 ID 无效")
		return
	}
	if strings.TrimSpace(query.Email) == "" && strings.TrimSpace(query.Username) == "" && query.ReviewID == 0 {
		respondBadRequest(c, "邮箱、用户名、点评 ID 至少填写一项")
		return
	}

	result, err := ctrl.query.FindUser(c.Request.Context(), application.AdminUserLookup{
		Email:    query.Email,
		Username: query.Username,
		ReviewID: query.ReviewID,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminUserController) ListAdmins(c *gin.Context) {
	result, err := ctrl.query.ListAdmins(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminUserController) SuspendUser(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}
	cmd := suspendUserCommand{}

	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	if cmd.Days <= 0 {
		respondBadRequest(c, "封禁天数必须大于 0")
		return
	}

	actor := auth.GetUserFromCtx(c.Request.Context())
	if actor == nil {
		respondUnauthorized(c)
		return
	}

	if err := ctrl.command.SuspendUserForDays(c.Request.Context(), actor, userID, cmd.Days); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AdminUserController) ClearSuspension(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}

	actor := auth.GetUserFromCtx(c.Request.Context())
	if actor == nil {
		respondUnauthorized(c)
		return
	}

	if err := ctrl.command.ClearSuspension(c.Request.Context(), actor, userID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AdminUserController) GrantAdmin(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	if actor == nil {
		respondUnauthorized(c)
		return
	}

	if err := ctrl.command.GrantAdmin(c.Request.Context(), actor, userID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AdminUserController) RevokeAdmin(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	if actor == nil {
		respondUnauthorized(c)
		return
	}

	if err := ctrl.command.RevokeAdmin(c.Request.Context(), actor, userID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AdminUserController) ResetPassword(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}
	var cmd resetPasswordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	if actor == nil {
		respondUnauthorized(c)
		return
	}

	if err := ctrl.command.ResetPassword(c.Request.Context(), actor, userID, cmd.Password); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func bindAdminUserID(c *gin.Context) (int, bool) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		respondBadRequest(c, "用户 ID 无效")
		return 0, false
	}
	return userID, true
}
