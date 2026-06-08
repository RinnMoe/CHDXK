package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/announcement"
	"jcourse/pkg/apperr"
)

type AnnouncementController struct {
	announcementQuery   *application.AnnouncementQueryService
	announcementService *announcement.Service
}

func NewAnnouncementController(
	announcementQuery *application.AnnouncementQueryService,
	announcementService *announcement.Service,
) *AnnouncementController {
	return &AnnouncementController{announcementQuery: announcementQuery, announcementService: announcementService}
}

func (ctrl *AnnouncementController) ListAnnouncements(c *gin.Context) {
	result, err := ctrl.announcementQuery.ListActiveAnnouncements(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AnnouncementController) ListAdminAnnouncements(c *gin.Context) {
	result, err := ctrl.announcementQuery.ListAdminAnnouncements(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AnnouncementController) CreateAnnouncement(c *gin.Context) {
	var cmd application.SaveAnnouncementCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	input, err := bindAnnouncementInput(cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	resultItem, err := ctrl.announcementService.Create(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	result := application.NewAnnouncementDTOFromDomain(resultItem)
	c.JSON(http.StatusCreated, result)
}

func (ctrl *AnnouncementController) UpdateAnnouncement(c *gin.Context) {
	announcementID, ok := bindAnnouncementID(c)
	if !ok {
		return
	}
	var cmd application.SaveAnnouncementCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	input, err := bindAnnouncementInput(cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	resultItem, err := ctrl.announcementService.Update(c.Request.Context(), announcementID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	result := application.NewAnnouncementDTOFromDomain(resultItem)
	c.JSON(http.StatusOK, result)
}

func (ctrl *AnnouncementController) DeleteAnnouncement(c *gin.Context) {
	announcementID, ok := bindAnnouncementID(c)
	if !ok {
		return
	}
	if err := ctrl.announcementService.Delete(c.Request.Context(), announcementID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func bindAnnouncementInput(cmd application.SaveAnnouncementCommand) (announcement.SaveInput, error) {
	showStart, err := parseAnnouncementTime(cmd.ShowStart)
	if err != nil {
		return announcement.SaveInput{}, apperr.BadRequest("公告开始时间无效")
	}
	showEnd, err := parseAnnouncementTime(cmd.ShowEnd)
	if err != nil {
		return announcement.SaveInput{}, apperr.BadRequest("公告结束时间无效")
	}
	return announcement.SaveInput{
		Title:     cmd.Title,
		Body:      cmd.Body,
		Priority:  cmd.Priority,
		ShowStart: showStart,
		ShowEnd:   showEnd,
		LinkURL:   cmd.LinkURL,
		LinkTitle: cmd.LinkTitle,
	}, nil
}

func parseAnnouncementTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func bindAnnouncementID(c *gin.Context) (int, bool) {
	announcementID, err := strconv.Atoi(c.Param("announcementID"))
	if err != nil || announcementID <= 0 {
		respondBadRequest(c, "公告 ID 无效")
		return 0, false
	}
	return announcementID, true
}
