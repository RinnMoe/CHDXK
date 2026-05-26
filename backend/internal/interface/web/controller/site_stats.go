package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type SiteStatsController struct {
	query *application.SiteStatsQueryService
}

func NewSiteStatsController(query *application.SiteStatsQueryService) *SiteStatsController {
	return &SiteStatsController{query: query}
}

func (ctrl *SiteStatsController) GetByDate(c *gin.Context) {
	dateStr := c.Param("date")
	result, err := ctrl.query.GetByDateString(c.Request.Context(), dateStr)
	if err != nil {
		respondError(c, err)
		return
	}
	if result == nil {
		respondNotFound(c, "站点每日统计不存在")
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *SiteStatsController) ListDaily(c *gin.Context) {
	var f application.SiteDailyStatListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}

	result, err := ctrl.query.ListDaily(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
