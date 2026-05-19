package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type TeacherController struct {
	teacherQuery *application.TeacherQueryService
	courseQuery  *application.CourseQueryService
}

func NewTeacherController(teacherQuery *application.TeacherQueryService, courseQuery *application.CourseQueryService) *TeacherController {
	return &TeacherController{teacherQuery: teacherQuery, courseQuery: courseQuery}
}

func (ctrl *TeacherController) ListTeachers(c *gin.Context) {
	var f application.TeacherListFilter
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

	result, err := ctrl.teacherQuery.ListTeachers(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *TeacherController) ListTeacherCourses(c *gin.Context) {
	teacherIDStr := c.Param("teacherID")
	teacherID, err := strconv.Atoi(teacherIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher ID"})
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

	result, err := ctrl.courseQuery.ListTeacherCourses(c.Request.Context(), teacherID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
