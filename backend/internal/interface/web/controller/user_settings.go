package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
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
		respondUnauthorized(c)
		return
	}

	result, err := ctrl.query.Get(c.Request.Context(), u.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *UserSettingsController) UpdateMySettings(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var cmd application.UpdateUserSettingsCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	result, err := ctrl.command.Update(c.Request.Context(), u.ID, cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
