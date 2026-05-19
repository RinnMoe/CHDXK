package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type TeacherController struct {
	query *application.TeacherQueryService
}

func NewTeacherController(query *application.TeacherQueryService) *TeacherController {
	return &TeacherController{query: query}
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

	result, err := ctrl.query.ListTeachers(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
