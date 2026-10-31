package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenAdminController maintains halls and stores. Authorization is enforced
// by RequireAdmin at the router level, matching the other admin groups.
type CanteenAdminController struct {
	admin   *application.CanteenAdminService
	catalog *application.CanteenCatalogQueryService
}

func NewCanteenAdminController(
	admin *application.CanteenAdminService,
	catalog *application.CanteenCatalogQueryService,
) *CanteenAdminController {
	return &CanteenAdminController{admin: admin, catalog: catalog}
}

type canteenHallRequest struct {
	CampusKey       string   `json:"campus_key"`
	Name            string   `json:"name"`
	ShortName       string   `json:"short_name"`
	Address         string   `json:"address"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	ExternalPlaceID string   `json:"external_place_id"`
	Status          string   `json:"status"`
	DisplayOrder    int      `json:"display_order"`
}

func (r canteenHallRequest) command() application.CanteenHallCommand {
	return application.CanteenHallCommand{
		CampusKey:       r.CampusKey,
		Name:            r.Name,
		ShortName:       r.ShortName,
		Address:         r.Address,
		Latitude:        r.Latitude,
		Longitude:       r.Longitude,
		ExternalPlaceID: r.ExternalPlaceID,
		Status:          canteen.HallStatus(r.Status),
		DisplayOrder:    r.DisplayOrder,
	}
}

type canteenStoreRequest struct {
	HallID       int      `json:"hall_id"`
	Name         string   `json:"name"`
	Floor        string   `json:"floor"`
	Categories   []string `json:"categories"`
	Status       string   `json:"status"`
	CoverMediaID *int     `json:"cover_media_id"`
}

func (r canteenStoreRequest) command() application.CanteenStoreCommand {
	return application.CanteenStoreCommand{
		HallID:       r.HallID,
		Name:         r.Name,
		Floor:        r.Floor,
		Categories:   r.Categories,
		Status:       canteen.StoreStatus(r.Status),
		CoverMediaID: r.CoverMediaID,
	}
}

type canteenStoreStatusRequest struct {
	Status string `json:"status"`
}

type canteenStoreMergeRequest struct {
	TargetID int `json:"target_id"`
}

func (c *CanteenAdminController) CreateHall(ctx *gin.Context) {
	var request canteenHallRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	hall, err := c.admin.CreateHall(ctx.Request.Context(), currentActorID(ctx), request.command())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": hall.ID, "name": hall.Name, "status": hall.Status})
}

func (c *CanteenAdminController) UpdateHall(ctx *gin.Context) {
	hallID, err := strconv.Atoi(ctx.Param("hallID"))
	if err != nil || hallID <= 0 {
		respondBadRequest(ctx, "食堂 ID 无效")
		return
	}
	var request canteenHallRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	hall, err := c.admin.UpdateHall(ctx.Request.Context(), currentActorID(ctx), hallID, request.command())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": hall.ID, "name": hall.Name, "status": hall.Status})
}

func (c *CanteenAdminController) ListStores(ctx *gin.Context) {
	var query canteenStoreListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		respondBindError(ctx, err)
		return
	}
	status := canteen.StoreStatus(ctx.Query("status"))
	if status != "" && !status.Valid() {
		respondBadRequest(ctx, "店铺状态无效")
		return
	}
	result, err := c.catalog.ListStoresForAdmin(ctx.Request.Context(), canteen.StoreListFilter{
		Q:         query.Q,
		CampusKey: query.Campus,
		HallID:    query.HallID,
		Category:  query.Category,
		Status:    status,
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

func (c *CanteenAdminController) CreateStore(ctx *gin.Context) {
	var request canteenStoreRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	store, err := c.admin.CreateStore(ctx.Request.Context(), currentActorID(ctx), request.command())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": store.ID, "name": store.Name, "status": store.Status})
}

func (c *CanteenAdminController) UpdateStore(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var request canteenStoreRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	store, err := c.admin.UpdateStore(ctx.Request.Context(), currentActorID(ctx), storeID, request.command())
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": store.ID, "name": store.Name, "status": store.Status})
}

func (c *CanteenAdminController) UpdateStoreStatus(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var request canteenStoreStatusRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	store, err := c.admin.UpdateStoreStatus(ctx.Request.Context(), currentActorID(ctx), storeID, canteen.StoreStatus(request.Status))
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": store.ID, "status": store.Status})
}

func (c *CanteenAdminController) MergeStore(ctx *gin.Context) {
	storeID, err := strconv.Atoi(ctx.Param("storeID"))
	if err != nil || storeID <= 0 {
		respondBadRequest(ctx, "店铺 ID 无效")
		return
	}
	var request canteenStoreMergeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBindError(ctx, err)
		return
	}
	store, err := c.admin.MergeStore(ctx.Request.Context(), currentActorID(ctx), storeID, request.TargetID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": store.ID, "status": store.Status, "merged_into": store.MergedIntoStoreID})
}

func currentActorID(ctx *gin.Context) int {
	if user := auth.GetUserFromCtx(ctx.Request.Context()); user != nil {
		return user.ID
	}
	return 0
}
