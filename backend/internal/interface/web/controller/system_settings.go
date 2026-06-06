package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

type SystemSettingsController struct {
	query   *application.SystemSettingsQueryService
	command *application.SystemSettingsCommandService
}

func NewSystemSettingsController(query *application.SystemSettingsQueryService, command *application.SystemSettingsCommandService) *SystemSettingsController {
	return &SystemSettingsController{query: query, command: command}
}

func (ctrl *SystemSettingsController) List(c *gin.Context) {
	settings, err := ctrl.query.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (ctrl *SystemSettingsController) Update(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var cmd application.UpdateSystemSettingCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	setting, err := ctrl.command.Update(c.Request.Context(), u, c.Param("key"), cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, setting)
}
