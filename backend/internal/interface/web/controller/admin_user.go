package controller

import (
	"net/http"
	"strconv"

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

func NewAdminUserController(
	query *application.AdminUserQueryService,
	command *application.AdminUserCommandService,
) *AdminUserController {
	return &AdminUserController{query: query, command: command}
}

func (ctrl *AdminUserController) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		respondBadRequest(c, "邮箱不能为空")
		return
	}

	result, err := ctrl.query.FindByEmail(c.Request.Context(), email)
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
	cmd := suspendUserCommand{Days: 30}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&cmd); err != nil {
			respondBindError(c, err)
			return
		}
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
