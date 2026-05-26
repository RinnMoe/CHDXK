package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/audit"
)

type AuditLogController struct {
	query *application.AuditLogQueryService
}

func NewAuditLogController(query *application.AuditLogQueryService) *AuditLogController {
	return &AuditLogController{query: query}
}

func (ctrl *AuditLogController) ListAuditLogs(c *gin.Context) {
	var f application.AuditLogListFilter
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

	result, err := ctrl.query.List(c.Request.Context(), f)
	if err != nil {
		if errors.Is(err, audit.ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
