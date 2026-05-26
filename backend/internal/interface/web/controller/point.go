package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

type PointController struct {
	query   *application.PointQueryService
	command *application.PointCommandService
}

func NewPointController(query *application.PointQueryService, command *application.PointCommandService) *PointController {
	return &PointController{query: query, command: command}
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
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 20
	}

	result, err := ctrl.query.GetUserPoints(c.Request.Context(), userID, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *PointController) CreateTransfer(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var cmd application.CreatePointTransferCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	transfer, err := ctrl.command.CreateTransfer(c.Request.Context(), u, cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

func (ctrl *PointController) PreviewTransfer(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var params application.PreviewTransferParams
	if err := c.ShouldBindJSON(&params); err != nil {
		respondBindError(c, err)
		return
	}

	preview, err := ctrl.query.PreviewTransfer(c.Request.Context(), u.ID, params)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (ctrl *PointController) GetPointsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		respondBadRequest(c, "邮箱不能为空")
		return
	}

	total, err := ctrl.query.GetUserPointsByEmail(c.Request.Context(), email)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"email": email, "total": total})
}
