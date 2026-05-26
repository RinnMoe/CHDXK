package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if u.ID != userID && !u.IsAdmin() {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var f application.PointRecordListFilter
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

	result, err := ctrl.query.GetUserPoints(c.Request.Context(), userID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *PointController) CreateTransfer(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var cmd application.CreatePointTransferCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := ctrl.command.CreateTransfer(c.Request.Context(), u, cmd)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrPointTransferInvalidAmount),
			errors.Is(err, application.ErrPointTransferInvalidFeePayer),
			errors.Is(err, application.ErrPointTransferSelf),
			errors.Is(err, application.ErrPointTransferRecipientAmountSmall),
			errors.Is(err, identity.ErrEmailNotAllowed):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, application.ErrPointTransferRecipientNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, point.ErrInsufficientBalance):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

func (ctrl *PointController) PreviewTransfer(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var params application.PreviewTransferParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preview, err := ctrl.query.PreviewTransfer(c.Request.Context(), u.ID, params)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrPointTransferInvalidAmount),
			errors.Is(err, application.ErrPointTransferInvalidFeePayer),
			errors.Is(err, application.ErrPointTransferRecipientAmountSmall):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (ctrl *PointController) GetPointsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	total, err := ctrl.query.GetUserPointsByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, application.ErrPointUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"email": email, "total": total})
}
