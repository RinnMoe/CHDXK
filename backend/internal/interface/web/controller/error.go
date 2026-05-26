package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/pkg/apperr"
)

func respondError(c *gin.Context, err error) {
	if appErr, ok := errors.AsType[*apperr.AppError](err); ok {
		c.JSON(appErr.StatusCode, gin.H{"error": appErr.Msg})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
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
