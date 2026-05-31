package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
	"jcourse/pkg/logx"
)

type AccountController struct {
	command *application.AccountCommandService
	query   *application.AccountQueryService
}

func NewAccountController(command *application.AccountCommandService, query *application.AccountQueryService) *AccountController {
	return &AccountController{command: command, query: query}
}

func (ctrl *AccountController) SendRegisterCode(c *gin.Context) {
	var cmd application.SendRegisterCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "send_register_code", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendRegisterCode(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "send_register_code", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) Register(c *gin.Context) {
	var cmd application.RegisterCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "register", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Register(c.Request.Context(), cmd)
	if err != nil {
		logAccountError(c, "register", cmd.Email, err)
		respondError(c, err)
		return
	}
	authHash, err := ctrl.command.SessionAuthHash(c.Request.Context(), u.ID)
	if err != nil {
		logAccountError(c, "register_session_auth_hash", cmd.Email, err)
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUser(c, u.ID, authHash); err != nil {
		logAccountError(c, "register_set_session", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (ctrl *AccountController) Login(c *gin.Context) {
	var cmd application.LoginCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "login", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Login(c.Request.Context(), cmd)
	if err != nil {
		logAccountError(c, "login", cmd.Email, err)
		respondError(c, err)
		return
	}
	authHash, err := ctrl.command.SessionAuthHash(c.Request.Context(), u.ID)
	if err != nil {
		logAccountError(c, "login_session_auth_hash", cmd.Email, err)
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUser(c, u.ID, authHash); err != nil {
		logAccountError(c, "login_set_session", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (ctrl *AccountController) Logout(c *gin.Context) {
	if err := middleware.ClearSession(c); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) Me(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}
	dto, err := ctrl.query.CurrentUser(c.Request.Context(), u)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto)
}

func (ctrl *AccountController) SendResetCode(c *gin.Context) {
	var cmd application.SendResetCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "send_reset_code", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendResetCode(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "send_reset_code", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) ResetPassword(c *gin.Context) {
	var cmd application.ResetPasswordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "reset_password", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.ResetPassword(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "reset_password", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func logAccountError(c *gin.Context, action, email string, err error) {
	logx.Warn(c.Request.Context(), "account command failed", "action", action, "email", email, "err", err)
}
