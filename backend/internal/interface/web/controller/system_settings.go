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

func (ctrl *SystemSettingsController) Copy(c *gin.Context) {
	values, err := ctrl.query.PublicCopy(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, values)
}

func (ctrl *SystemSettingsController) Pages(c *gin.Context) {
	pages, err := ctrl.query.Pages(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pages)
}

func (ctrl *SystemSettingsController) Manifest(c *gin.Context) {
	values, err := ctrl.query.PublicCopy(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	palette, err := ctrl.query.Theme(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	name := values["copy.user.config.brand.743e9916a4b2"]
	icons := []gin.H{{"src": "/pwa-192x192.png", "sizes": "192x192", "type": "image/png"}, {"src": "/pwa-512x512.png", "sizes": "512x512", "type": "image/png"}, {"src": "/pwa-512x512.png", "sizes": "512x512", "type": "image/png", "purpose": "maskable"}}
	shortcuts := make([]gin.H, 0, 3)
	for _, item := range []struct{ key, path string }{{"course", "/course"}, {"review", "/review"}, {"mine", "/review/mine"}} {
		shortcuts = append(shortcuts, gin.H{"name": values["copy.pwa."+item.key+"_name"], "description": values["copy.pwa."+item.key+"_description"], "url": item.path, "icons": icons[:1]})
	}
	c.Header("Content-Type", "application/manifest+json")
	c.JSON(http.StatusOK, gin.H{"id": "/", "name": name, "short_name": name, "description": values["copy.pwa.description"], "start_url": "/", "scope": "/", "display": "standalone", "theme_color": palette["light"]["primary"], "background_color": palette["light"]["background"], "lang": "zh-CN", "categories": []string{"education", "productivity"}, "icons": icons, "shortcuts": shortcuts})
}

func (ctrl *SystemSettingsController) ListAdmin(c *gin.Context) {
	settings, err := ctrl.query.ListAdmin(c.Request.Context())
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

func (ctrl *SystemSettingsController) Theme(c *gin.Context) {
	palette, err := ctrl.query.Theme(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, palette)
}
