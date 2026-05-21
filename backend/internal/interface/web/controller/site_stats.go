package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jcourse/internal/application"
)

type SiteStatsController struct {
	query *application.SiteStatsQueryService
}

func NewSiteStatsController(query *application.SiteStatsQueryService) *SiteStatsController {
	return &SiteStatsController{query: query}
}

func (ctrl *SiteStatsController) GetYesterday(c *gin.Context) {
	result, err := ctrl.query.GetYesterday(c.Request.Context())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "site daily stats not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *SiteStatsController) ListDaily(c *gin.Context) {
	var f application.SiteDailyStatListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		if errors.Is(err, application.ErrInvalidDateRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
