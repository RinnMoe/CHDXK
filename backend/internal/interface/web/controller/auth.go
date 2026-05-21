package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type AuthController struct {
	command *application.AuthCommandService
}

func NewAuthController(command *application.AuthCommandService) *AuthController {
	return &AuthController{command: command}
}

func (ctrl *AuthController) SendRegisterCode(c *gin.Context) {
	var cmd application.SendRegisterCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.command.SendRegisterCode(c.Request.Context(), cmd); err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailNotAllowed):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrVerificationTooSoon):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var cmd application.RegisterCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := ctrl.command.Register(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailNotAllowed), errors.Is(err, auth.ErrVerificationCodeInvalid), errors.Is(err, auth.ErrPasswordRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if err := middleware.SetSessionUserID(c, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var cmd application.LoginCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := ctrl.command.Login(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrLoginLocked) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrUserSuspended) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := middleware.SetSessionUserID(c, u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	if err := middleware.ClearSession(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AuthController) SendResetCode(c *gin.Context) {
	var cmd application.SendResetCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.command.SendResetCode(c.Request.Context(), cmd); err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailNotAllowed):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrVerificationTooSoon):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var cmd application.ResetPasswordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.command.ResetPassword(c.Request.Context(), cmd); err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailNotAllowed), errors.Is(err, auth.ErrVerificationCodeInvalid), errors.Is(err, auth.ErrPasswordRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, auth.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
