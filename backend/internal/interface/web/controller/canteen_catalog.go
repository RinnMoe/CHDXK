package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenCatalogController exposes the public canteen directory.
type CanteenCatalogController struct {
	query *application.CanteenCatalogQueryService
}

func NewCanteenCatalogController(query *application.CanteenCatalogQueryService) *CanteenCatalogController {
	return &CanteenCatalogController{query: query}
}

type canteenStoreListQuery struct {
	Q        string `form:"q"`
	Campus   string `form:"campus"`
	HallID   int    `form:"hall_id"`
	Category string `form:"category"`
	Sort     string `form:"sort"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

func (c *CanteenCatalogController) Home(ctx *gin.Context) {
	result, err := c.query.Home(ctx.Request.Context())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *CanteenCatalogController) Halls(ctx *gin.Context) {
	result, err := c.query.ListHalls(ctx.Request.Context())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": result})
}

func (c *CanteenCatalogController) Stores(ctx *gin.Context) {
	var query canteenStoreListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		respondBindError(ctx, err)
		return
	}
	result, err := c.query.ListStores(ctx.Request.Context(), canteen.StoreListFilter{
		Q:         query.Q,
		CampusKey: query.Campus,
		HallID:    query.HallID,
		Category:  query.Category,
		Sort:      query.Sort,
		Page:      query.Page,
		PageSize:  query.PageSize,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *CanteenCatalogController) Rankings(ctx *gin.Context) {
	var query canteenStoreListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		respondBindError(ctx, err)
		return
	}
	result, err := c.query.ListRankings(ctx.Request.Context(), canteen.RankingFilter{
		CampusKey: query.Campus,
		HallID:    query.HallID,
		Sort:      query.Sort,
		Page:      query.Page,
		PageSize:  query.PageSize,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *CanteenCatalogController) Store(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	viewerID := 0
	if user := auth.GetUserFromCtx(ctx.Request.Context()); user != nil {
		viewerID = user.ID
	}
	result, err := c.query.GetStore(ctx.Request.Context(), storeID, viewerID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, result)
}
