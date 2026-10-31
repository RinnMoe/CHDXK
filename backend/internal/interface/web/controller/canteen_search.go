package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

// CanteenSearchController exposes grouped canteen + dish search.
type CanteenSearchController struct {
	query *application.CanteenSearchQueryService
}

func NewCanteenSearchController(query *application.CanteenSearchQueryService) *CanteenSearchController {
	return &CanteenSearchController{query: query}
}

func (c *CanteenSearchController) Search(ctx *gin.Context) {
	result, err := c.query.Search(ctx.Request.Context(), ctx.Query("q"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
