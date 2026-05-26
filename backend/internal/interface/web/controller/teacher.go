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

func (ctrl *TeacherController) GetTeacherFilters(c *gin.Context) {
	result, err := ctrl.teacherQuery.GetTeacherFilters(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *TeacherController) ListTeachers(c *gin.Context) {
	var f application.TeacherListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.teacherQuery.ListTeachers(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *TeacherController) GetTeacher(c *gin.Context) {
	teacherIDStr := c.Param("teacherID")
	teacherID, err := strconv.Atoi(teacherIDStr)
	if err != nil {
		respondBadRequest(c, "教师 ID 无效")
		return
	}

	result, err := ctrl.teacherQuery.GetTeacher(c.Request.Context(), teacherID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *TeacherController) ListTeacherCourses(c *gin.Context) {
	teacherIDStr := c.Param("teacherID")
	teacherID, err := strconv.Atoi(teacherIDStr)
	if err != nil {
		respondBadRequest(c, "教师 ID 无效")
		return
	}

	var f application.CourseListFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		respondBindError(c, err)
		return
	}
	normalizePagination(&f.Page, &f.PageSize)

	result, err := ctrl.courseQuery.ListTeacherCourses(c.Request.Context(), teacherID, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
