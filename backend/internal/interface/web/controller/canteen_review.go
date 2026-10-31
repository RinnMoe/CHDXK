package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenReviewController exposes visit reviews and votes.
type CanteenReviewController struct {
	command *application.CanteenReviewCommandService
	query   *application.CanteenReviewQueryService
}

func NewCanteenReviewController(
	command *application.CanteenReviewCommandService,
	query *application.CanteenReviewQueryService,
) *CanteenReviewController {
	return &CanteenReviewController{command: command, query: query}
}

type canteenReviewBody struct {
	VisitedAt       string   `json:"visited_at"`
	Taste           int      `json:"taste"`
	Value           int      `json:"value"`
	Queue           int      `json:"queue"`
	Hygiene         int      `json:"hygiene"`
	Service         int      `json:"service"`
	Content         string   `json:"content"`
	Tags            []string `json:"tags"`
	ClientRequestID string   `json:"client_request_id"`
	MediaIDs        []int    `json:"media_ids"`
}

type canteenReviewUpdateBody struct {
	VisitedAt         string   `json:"visited_at"`
	Taste             int      `json:"taste"`
	Value             int      `json:"value"`
	Queue             int      `json:"queue"`
	Hygiene           int      `json:"hygiene"`
	Service           int      `json:"service"`
	Content           string   `json:"content"`
	Tags              []string `json:"tags"`
	ExpectedUpdatedAt string   `json:"expected_updated_at"`
}

type canteenVoteBody struct {
	VoteType int `json:"vote_type"`
}

type canteenHideBody struct {
	Reason string `json:"reason"`
}

type canteenReviewListQuery struct {
	Order     string `form:"order"`
	Filter    string `form:"filter"`
	WithMedia bool   `form:"with_media"`
	History   bool   `form:"history"`
	Cursor    string `form:"cursor"`
	Limit     int    `form:"limit"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

func parseCanteenTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}

func (c *CanteenReviewController) Create(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var body canteenReviewBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	visitedAt, err := parseCanteenTime(body.VisitedAt)
	if err != nil {
		respondBadRequest(ctx, "到店时间无效")
		return
	}
	requestID := strings.TrimSpace(ctx.GetHeader("Idempotency-Key"))
	if requestID == "" {
		requestID = strings.TrimSpace(body.ClientRequestID)
	}
	if requestID == "" {
		respondBadRequest(ctx, "缺少幂等请求标识")
		return
	}
	if _, err := uuid.Parse(requestID); err != nil {
		respondBadRequest(ctx, "幂等请求标识格式无效")
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	review, created, err := c.command.Create(ctx.Request.Context(), user, application.CreateCanteenReviewCommand{
		StoreID:         storeID,
		VisitedAt:       visitedAt,
		Taste:           body.Taste,
		Value:           body.Value,
		Queue:           body.Queue,
		Hygiene:         body.Hygiene,
		Service:         body.Service,
		Content:         body.Content,
		Tags:            body.Tags,
		ClientRequestID: requestID,
		MediaIDs:        body.MediaIDs,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	dto, err := c.query.GetReview(ctx.Request.Context(), review.ID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"created": created, "review": dto})
}

func (c *CanteenReviewController) Update(ctx *gin.Context) {
	reviewID, err := strconv.Atoi(ctx.Param("reviewID"))
	if err != nil || reviewID <= 0 {
		respondBadRequest(ctx, "评价 ID 无效")
		return
	}
	var body canteenReviewUpdateBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	visitedAt, err := parseCanteenTime(body.VisitedAt)
	if err != nil {
		respondBadRequest(ctx, "到店时间无效")
		return
	}
	expected, err := parseCanteenTime(body.ExpectedUpdatedAt)
	if err != nil {
		respondBadRequest(ctx, "版本时间无效")
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	review, err := c.command.Update(ctx.Request.Context(), user, application.UpdateCanteenReviewCommand{
		ReviewID:          reviewID,
		VisitedAt:         visitedAt,
		Taste:             body.Taste,
		Value:             body.Value,
		Queue:             body.Queue,
		Hygiene:           body.Hygiene,
		Service:           body.Service,
		Content:           body.Content,
		Tags:              body.Tags,
		ExpectedUpdatedAt: expected,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	dto, err := c.query.GetReview(ctx.Request.Context(), review.ID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"review": dto})
}

func (c *CanteenReviewController) Delete(ctx *gin.Context) {
	reviewID, err := strconv.Atoi(ctx.Param("reviewID"))
	if err != nil || reviewID <= 0 {
		respondBadRequest(ctx, "评价 ID 无效")
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	if err := c.command.Delete(ctx.Request.Context(), user, reviewID); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (c *CanteenReviewController) Vote(ctx *gin.Context) {
	reviewID, err := strconv.Atoi(ctx.Param("reviewID"))
	if err != nil || reviewID <= 0 {
		respondBadRequest(ctx, "评价 ID 无效")
		return
	}
	var body canteenVoteBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	if err := c.command.Vote(ctx.Request.Context(), user, reviewID, body.VoteType); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (c *CanteenReviewController) Hide(ctx *gin.Context) {
	reviewID, err := strconv.Atoi(ctx.Param("reviewID"))
	if err != nil || reviewID <= 0 {
		respondBadRequest(ctx, "评价 ID 无效")
		return
	}
	var body canteenHideBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	if err := c.command.Hide(ctx.Request.Context(), user, reviewID, body.Reason); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (c *CanteenReviewController) ListStoreReviews(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var query canteenReviewListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		respondBindError(ctx, err)
		return
	}
	viewerID := 0
	if user := auth.GetUserFromCtx(ctx.Request.Context()); user != nil {
		viewerID = user.ID
	}
	filter := canteen.ReviewListFilter{
		ViewerID:  viewerID,
		Order:     query.Order,
		WithMedia: query.WithMedia,
		History:   query.History,
		Page:      query.Page,
		PageSize:  query.PageSize,
	}
	if query.Limit > 0 {
		filter.PageSize = query.Limit
	}
	switch strings.ToLower(strings.TrimSpace(query.Filter)) {
	case "", "all":
	case "with_media":
		filter.WithMedia = true
	case "high":
		minimum := 4.0
		filter.MinScore = &minimum
	case "low":
		maximum := 2.0
		filter.MaxScore = &maximum
	default:
		respondBadRequest(ctx, "评价筛选参数无效")
		return
	}
	if query.Cursor != "" {
		cursor, err := application.DecodeCanteenReviewCursor(query.Cursor)
		if err != nil {
			respondError(ctx, err)
			return
		}
		filter.Cursor = cursor
	}
	page, err := c.query.ListStoreReviews(ctx.Request.Context(), storeID, filter)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, page)
}

func (c *CanteenReviewController) ListMyReviews(ctx *gin.Context) {
	var query canteenReviewListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	page, err := c.query.ListMyReviews(ctx.Request.Context(), user.ID, query.Page, query.PageSize)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, page)
}
