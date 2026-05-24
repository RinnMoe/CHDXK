package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/setting"
)

type UserSettingsController struct {
	query   *application.UserSettingsQueryService
	command *application.UserSettingsCommandService
}

func NewUserSettingsController(query *application.UserSettingsQueryService, command *application.UserSettingsCommandService) *UserSettingsController {
	return &UserSettingsController{query: query, command: command}
}

func (ctrl *UserSettingsController) GetMySettings(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := ctrl.query.Get(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *UserSettingsController) UpdateMySettings(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var cmd application.UpdateUserSettingsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := ctrl.command.Update(c.Request.Context(), u.ID, cmd)
	if err != nil {
		if errors.Is(err, setting.ErrInvalidCurrentSemester) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
