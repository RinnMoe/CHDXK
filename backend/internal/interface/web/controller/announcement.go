package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type AnnouncementController struct {
	announcementQuery *application.AnnouncementQueryService
}

func NewAnnouncementController(announcementQuery *application.AnnouncementQueryService) *AnnouncementController {
	return &AnnouncementController{announcementQuery: announcementQuery}
}

func (ctrl *AnnouncementController) ListAnnouncements(c *gin.Context) {
	result, err := ctrl.announcementQuery.ListActiveAnnouncements(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
