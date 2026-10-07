package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/domain/auth"
	"jcourse/pkg/apperr"
	"jcourse/pkg/sitecopy"
)

func respondError(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*apperr.AppError](err); ok {
		body := gin.H{"error": sitecopy.Text(c.Request.Context(), appErr.Msg)}
		switch {
		case errors.Is(err, auth.ErrPasswordRequired):
			body["code"] = "password_change_required"
		case errors.Is(err, auth.ErrUserSuspended):
			body["code"] = "account_suspended"
		case errors.Is(err, auth.ErrMuted):
			body["code"] = "muted"
		}
		c.JSON(appErr.StatusCode, body)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": sitecopy.Text(c.Request.Context(), "服务器内部错误")})
}

func respondBadRequest(c *gin.Context, msg string) {
	respondError(c, apperr.BadRequest(msg))
}

func respondBindError(c *gin.Context, err error) {
	respondError(c, apperr.BadRequest("请求参数无效"))
}

func respondUnauthorized(c *gin.Context) {
	respondError(c, apperr.ErrUnauthorized)
}

func respondNotFound(c *gin.Context, msg string) {
	respondError(c, apperr.NotFound(msg))
}
