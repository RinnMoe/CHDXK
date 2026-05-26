package controller

import (
	"errors"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := ctrl.query.ListMyEnrollments(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseEnrollmentController) CreateEnrollment(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	var req struct {
		Semester string `json:"semester"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ctrl.command.CreateEnrollment(c.Request.Context(), u.ID, application.CreateCourseEnrollmentCommand{
		CourseID: courseID,
		Semester: req.Semester,
	})
	if err != nil {
		switch {
		case errors.Is(err, application.ErrSemesterRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "semester required"})
		case errors.Is(err, application.ErrOfferedCourseNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "offered course not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *CourseEnrollmentController) DeleteEnrollment(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	enrollmentID, err := strconv.Atoi(c.Param("enrollmentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id"})
		return
	}

	if err := ctrl.command.DeleteEnrollment(c.Request.Context(), u.ID, enrollmentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
