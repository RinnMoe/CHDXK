package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/teacher"
	"net/http"
	"strconv"
)

type TeacherProfileController struct {
	service *application.TeacherProfileService
}

func NewTeacherProfileController(service *application.TeacherProfileService) *TeacherProfileController {
	return &TeacherProfileController{service: service}
}
func (c *TeacherProfileController) Status(g *gin.Context) { g.JSON(http.StatusOK, c.service.Status()) }
func (c *TeacherProfileController) Start(g *gin.Context) {
	result, err := c.service.Start()
	if err != nil {
		respondError(g, err)
		return
	}
	g.JSON(http.StatusAccepted, result)
}
func (c *TeacherProfileController) Cancel(g *gin.Context) {
	c.service.Cancel()
	g.JSON(http.StatusOK, c.service.Status())
}
func (c *TeacherProfileController) Matches(g *gin.Context) {
	result, err := c.service.Matches(g.Request.Context())
	if err != nil {
		respondError(g, err)
		return
	}
	g.JSON(http.StatusOK, result)
}
func (c *TeacherProfileController) Bind(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("teacherID"))
	if err != nil || id < 1 {
		respondBadRequest(g, "教师编号无效")
		return
	}
	var cmd struct {
		URL string `json:"url" binding:"required"`
	}
	if err := g.ShouldBindJSON(&cmd); err != nil {
		respondBindError(g, err)
		return
	}
	if err := c.service.Bind(g.Request.Context(), id, cmd.URL); err != nil {
		respondError(g, err)
		return
	}
	g.Status(http.StatusNoContent)
}

func (c *TeacherProfileController) Teachers(g *gin.Context) {
	var f application.TeacherListFilter
	if err := g.ShouldBindQuery(&f); err != nil {
		respondBindError(g, err)
		return
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	rows, total, err := c.service.Teachers(g.Request.Context(), teacher.TeacherFilter{Q: f.Q, Department: f.Department, Title: f.Title, Page: f.Page, PageSize: f.PageSize})
	if err != nil {
		respondError(g, err)
		return
	}
	g.JSON(http.StatusOK, gin.H{"items": rows, "total": total, "page": f.Page, "page_size": f.PageSize})
}
func (c *TeacherProfileController) Edit(g *gin.Context) {
	id, err := strconv.Atoi(g.Param("teacherID"))
	if err != nil || id < 1 {
		respondBadRequest(g, "教师编号无效")
		return
	}
	var p teacher.Profile
	if err := g.ShouldBindJSON(&p); err != nil {
		respondBindError(g, err)
		return
	}
	if err := c.service.Edit(g.Request.Context(), id, p); err != nil {
		respondError(g, err)
		return
	}
	g.Status(http.StatusNoContent)
}
