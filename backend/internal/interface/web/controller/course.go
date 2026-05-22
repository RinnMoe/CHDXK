package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

type CourseController struct {
	query   *application.CourseQueryService
	command *application.CourseCommandService
}

func NewCourseController(query *application.CourseQueryService, command *application.CourseCommandService) *CourseController {
	return &CourseController{query: query, command: command}
}

func (ctrl *CourseController) GetCourseFilters(c *gin.Context) {
	result, err := ctrl.query.GetCourseFilters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListCourses(c *gin.Context) {
	var f application.CourseListFilter
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

	result, err := ctrl.query.ListCourses(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListHotCourses(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	result, err := ctrl.query.ListHotCourses(c.Request.Context(), period)
	if err != nil {
		if errors.Is(err, course.ErrInvalidHotCoursePeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) GetCourse(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())

	result, err := ctrl.query.GetCourseDetail(c.Request.Context(), u, courseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) SetNotificationLevel(c *gin.Context) {
	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Level int `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Level < 0 || req.Level > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "level must be 0, 1, or 2"})
		return
	}

	if err := ctrl.command.SetNotificationLevel(c.Request.Context(), u.ID, courseID, course.NotificationLevel(req.Level)); err != nil {
		if errors.Is(err, application.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *CourseController) ListFollowedCourses(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var f application.CourseListFilter
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

	result, err := ctrl.query.ListCoursesByNotificationLevel(c.Request.Context(), u.ID, course.NotificationLevelFollow, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseController) ListIgnoredCourses(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var f application.CourseListFilter
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

	result, err := ctrl.query.ListCoursesByNotificationLevel(c.Request.Context(), u.ID, course.NotificationLevelIgnored, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
