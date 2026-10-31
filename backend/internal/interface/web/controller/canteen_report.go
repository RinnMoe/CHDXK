package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/canteen"
)

// CanteenReportController exposes user reports and the admin queue.
type CanteenReportController struct {
	command *application.CanteenReportCommandService
}

func NewCanteenReportController(command *application.CanteenReportCommandService) *CanteenReportController {
	return &CanteenReportController{command: command}
}

type canteenReportBody struct {
	SubjectType string `json:"subject_type"`
	SubjectID   int    `json:"subject_id"`
	ReasonCode  string `json:"reason_code"`
	Detail      string `json:"detail"`
}

type canteenReportResolveBody struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func (c *CanteenReportController) Create(ctx *gin.Context) {
	var body canteenReportBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	user := auth.GetUserFromCtx(ctx.Request.Context())
	report, err := c.command.Create(ctx.Request.Context(), user, application.CreateCanteenReportCommand{
		SubjectType: canteen.ReportSubjectType(body.SubjectType),
		SubjectID:   body.SubjectID,
		Reason:      canteen.ReportReason(body.ReasonCode),
		Detail:      body.Detail,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": report.ID, "status": report.Status})
}

func (c *CanteenReportController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))
	items, total, page, pageSize, err := c.command.List(ctx.Request.Context(), auth.GetUserFromCtx(ctx.Request.Context()), canteen.ReportFilter{
		Status:   canteen.ReportStatus(ctx.Query("status")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (c *CanteenReportController) Resolve(ctx *gin.Context) {
	reportID, err := strconv.Atoi(ctx.Param("reportID"))
	if err != nil || reportID <= 0 {
		respondBadRequest(ctx, "举报 ID 无效")
		return
	}
	var body canteenReportResolveBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBindError(ctx, err)
		return
	}
	if err := c.command.Resolve(ctx.Request.Context(), auth.GetUserFromCtx(ctx.Request.Context()), reportID, canteen.ReportStatus(body.Status), body.Note); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}
