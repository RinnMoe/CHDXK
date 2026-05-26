package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
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
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendRegisterCode(c.Request.Context(), cmd); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) Register(c *gin.Context) {
	var cmd application.RegisterCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Register(c.Request.Context(), cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUserID(c, u.ID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (ctrl *AccountController) Login(c *gin.Context) {
	var cmd application.LoginCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Login(c.Request.Context(), cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUserID(c, u.ID); err != nil {
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
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendResetCode(c.Request.Context(), cmd); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) ResetPassword(c *gin.Context) {
	var cmd application.ResetPasswordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.ResetPassword(c.Request.Context(), cmd); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
