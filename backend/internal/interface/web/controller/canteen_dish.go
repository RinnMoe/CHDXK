package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenDishController exposes dish lists, suggestions, contributions and
// dish ratings.
type CanteenDishController struct {
	command *application.CanteenDishCommandService
	query   *application.CanteenDishQueryService
	media   *application.CanteenMediaQueryService
}

func NewCanteenDishController(
	command *application.CanteenDishCommandService,
	query *application.CanteenDishQueryService,
	media *application.CanteenMediaQueryService,
) *CanteenDishController {
	return &CanteenDishController{command: command, query: query, media: media}
}

type canteenDishBody struct {
	Name string `json:"name"`
}

type canteenDishReviewBody struct {
	ReviewEventID *int   `json:"review_event_id"`
	Taste         int    `json:"taste"`
	Portion       int    `json:"portion"`
	Value         int    `json:"value"`
	Content       string `json:"content"`
	MediaIDs      []int  `json:"media_ids"`
}

type canteenDishStatusBody struct {
	Status string `json:"status"`
}

type canteenDishMergeBody struct {
	TargetID int `json:"target_id"`
}

func (c *CanteenDishController) ListStoreDishes(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	items, err := c.query.ListStoreDishes(ctx.Request.Context(), storeID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (c *CanteenDishController) Suggest(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Query("store_id"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	items, err := c.query.Suggest(ctx.Request.Context(), storeID, ctx.Query("q"))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (c *CanteenDishController) CreateDish(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var body canteenDishBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	dish, existing, err := c.command.CreateDish(ctx.Request.Context(), user, application.CreateCanteenDishCommand{
		StoreID: storeID,
		Name:    body.Name,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":       dish.ID,
		"name":     dish.Name,
		"status":   dish.Status,
		"existing": existing,
	})
}

func (c *CanteenDishController) CreateDishReview(ctx *gin.Context) {
	dishID, err := strconv.Atoi(ctx.Param("dishID"))
	if err != nil || dishID <= 0 {
		respondBadRequest(ctx, "菜品 ID 无效")
		return
	}
	var body canteenDishReviewBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	review, err := c.command.CreateDishReview(ctx.Request.Context(), user, application.CreateCanteenDishReviewCommand{
		DishID:        dishID,
		ReviewEventID: body.ReviewEventID,
		Taste:         body.Taste,
		Portion:       body.Portion,
		Value:         body.Value,
		Content:       body.Content,
		MediaIDs:      body.MediaIDs,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":            review.ID,
		"dish_id":       review.DishID,
		"overall_score": review.OverallScore,
	})
}

func (c *CanteenDishController) GetDishDetail(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	dishID, err := strconv.Atoi(ctx.Param("dishID"))
	if err != nil || dishID <= 0 {
		respondBadRequest(ctx, "菜品 ID 无效")
		return
	}
	detail, err := c.query.GetDishDetail(ctx.Request.Context(), storeID, dishID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, detail)
}

func (c *CanteenDishController) ListAdminDishes(ctx *gin.Context) {
	storeID, _ := strconv.Atoi(ctx.Query("store_id"))
	filter := canteen.DishListFilter{
		StoreID: storeID,
		Status:  canteen.DishStatus(ctx.Query("status")),
		Q:       ctx.Query("q"),
		Limit:   100,
	}
	items, err := c.query.ListDishesForAdmin(ctx.Request.Context(), filter)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (c *CanteenDishController) UpdateDishStatus(ctx *gin.Context) {
	dishID, err := strconv.Atoi(ctx.Param("dishID"))
	if err != nil || dishID <= 0 {
		respondBadRequest(ctx, "菜品 ID 无效")
		return
	}
	var body canteenDishStatusBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	dish, err := c.command.UpdateDishStatus(ctx.Request.Context(), auth.GetUserFromCtx(ctx.Request.Context()), dishID, canteen.DishStatus(body.Status))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": dish.ID, "status": dish.Status})
}

func (c *CanteenDishController) MergeDish(ctx *gin.Context) {
	dishID, err := strconv.Atoi(ctx.Param("dishID"))
	if err != nil || dishID <= 0 {
		respondBadRequest(ctx, "菜品 ID 无效")
		return
	}
	var body canteenDishMergeBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	if err := c.command.MergeDish(ctx.Request.Context(), auth.GetUserFromCtx(ctx.Request.Context()), dishID, body.TargetID); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (c *CanteenDishController) Contributions(ctx *gin.Context) {
	user := auth.GetUserFromCtx(ctx.Request.Context())
	dishes, total, page, pageSize, err := c.query.ListMyDishReviews(ctx.Request.Context(), user.ID, 1, 50)
	if err != nil {
		respondError(ctx, err)
		return
	}
	media, mediaTotal, _, _, err := c.media.ListOwnerAssets(ctx.Request.Context(), user.ID, 1, 50)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"dishes":      dishes,
		"media":       media,
		"dish_total":  total,
		"media_total": mediaTotal,
		"page":        page,
		"page_size":   pageSize,
	})
}
