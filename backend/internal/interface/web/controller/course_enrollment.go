package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

type CourseEnrollmentController struct {
	query   *application.CourseEnrollmentQueryService
	command *application.CourseCommandService
}

func NewCourseEnrollmentController(query *application.CourseEnrollmentQueryService, command *application.CourseCommandService) *CourseEnrollmentController {
	return &CourseEnrollmentController{query: query, command: command}
}

func (ctrl *CourseEnrollmentController) ListMyEnrollments(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	result, err := ctrl.query.ListMyEnrollments(c.Request.Context(), u.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseEnrollmentController) DeleteEnrollment(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	enrollmentID, err := strconv.Atoi(c.Param("enrollmentID"))
	if err != nil {
		respondBadRequest(c, "选课记录 ID 无效")
		return
	}

	if err := ctrl.command.DeleteEnrollment(c.Request.Context(), u.ID, enrollmentID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
