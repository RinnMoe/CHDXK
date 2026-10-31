package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenMediaController implements the proxy upload flow and image serving.
type CanteenMediaController struct {
	command *application.CanteenMediaCommandService
	query   *application.CanteenMediaQueryService
}

func NewCanteenMediaController(
	command *application.CanteenMediaCommandService,
	query *application.CanteenMediaQueryService,
) *CanteenMediaController {
	return &CanteenMediaController{command: command, query: query}
}

type canteenMediaIntentBody struct {
	Purpose string `json:"purpose"`
	MIME    string `json:"mime"`
	Bytes   int64  `json:"bytes"`
}

type canteenMediaReviewBody struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (c *CanteenMediaController) CreateIntent(ctx *gin.Context) {
	var body canteenMediaIntentBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	media, uploadURL, err := c.command.CreateIntent(ctx.Request.Context(), user, body.Purpose, body.MIME, body.Bytes)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"media_id":   media.ID,
		"expires_at": media.ExpiresAt,
		"upload": gin.H{
			"mode":   "proxy",
			"method": http.MethodPut,
			"url":    uploadURL,
			"headers": gin.H{
				"Content-Type": media.MIME,
			},
		},
	})
}

func (c *CanteenMediaController) PutContent(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("mediaID"))
	if err != nil || mediaID <= 0 {
		respondBadRequest(ctx, "图片 ID 无效")
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	media, err := c.command.PutContent(ctx.Request.Context(), user, mediaID, ctx.Request.Body)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, newCanteenMediaResponse(media))
}

func (c *CanteenMediaController) Complete(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("mediaID"))
	if err != nil || mediaID <= 0 {
		respondBadRequest(ctx, "图片 ID 无效")
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	media, err := c.command.Complete(ctx.Request.Context(), user, mediaID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, newCanteenMediaResponse(media))
}

func (c *CanteenMediaController) Get(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("mediaID"))
	if err != nil || mediaID <= 0 {
		respondBadRequest(ctx, "图片 ID 无效")
		return
	}
	viewer := auth.GetUserFromCtx(ctx.Request.Context())
	reader, media, err := c.command.Open(ctx.Request.Context(), viewer, mediaID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()
	ctx.Header("Content-Type", media.MIME)
	ctx.Header("X-Content-Type-Options", "nosniff")
	// Canteen media visibility depends on parent state that can change at any
	// time. Do not let shared caches retain a response that was public before a
	// review/dish/store was hidden or deleted.
	ctx.Header("Cache-Control", "private, no-store")
	ctx.Status(http.StatusOK)
	if _, err := io.Copy(ctx.Writer, reader); err != nil {
		return
	}
}

func (c *CanteenMediaController) ListAdminMedia(ctx *gin.Context) {
	status := canteen.MediaStatus(ctx.Query("status"))
	if status != "" && !status.Valid() {
		respondBadRequest(ctx, "图片状态无效")
		return
	}
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))
	items, total, page, pageSize, err := c.query.ListAdminAssets(ctx.Request.Context(), canteen.MediaFilter{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (c *CanteenMediaController) Review(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("mediaID"))
	if err != nil || mediaID <= 0 {
		respondBadRequest(ctx, "图片 ID 无效")
		return
	}
	var body canteenMediaReviewBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	reviewer := auth.GetUserFromCtx(ctx.Request.Context())
	if err := c.command.Approve(ctx.Request.Context(), reviewer, mediaID, canteen.MediaStatus(body.Status), body.Reason); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func newCanteenMediaResponse(media *canteen.Media) gin.H {
	if media == nil {
		return gin.H{}
	}
	return gin.H{
		"id":         media.ID,
		"url":        fmt.Sprintf("/api/canteen/media/%d", media.ID),
		"mime":       media.MIME,
		"byte_size":  media.ByteSize,
		"width":      media.Width,
		"height":     media.Height,
		"status":     media.Status,
		"created_at": media.CreatedAt,
	}
}
