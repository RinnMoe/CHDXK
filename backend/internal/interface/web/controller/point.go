package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type PointController struct {
	query *application.PointQueryService
}

func NewPointController(query *application.PointQueryService) *PointController {
	return &PointController{query: query}
}

func (ctrl *PointController) GetUserPoints(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		respondBadRequest(c, "用户 ID 无效")
		return
	}

	var f application.PointRecordListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.query.GetUserPoints(c.Request.Context(), userID, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
