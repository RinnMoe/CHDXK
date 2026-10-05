package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
)

type CourseAdminController struct {
	service *application.CourseManagementService
}

func NewCourseAdminController(service *application.CourseManagementService) *CourseAdminController {
	return &CourseAdminController{service: service}
}

func (ctrl *CourseAdminController) ListSemesters(c *gin.Context) {
	semesters, err := ctrl.service.ListSemesters(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, semesters)
}

func (ctrl *CourseAdminController) List(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		respondBadRequest(c, "页码无效")
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if err != nil {
		respondBadRequest(c, "每页数量无效")
		return
	}
	filter := course.ManagedOfferingFilter{
		Semester:   strings.TrimSpace(c.Query("semester")),
		Code:       strings.TrimSpace(c.Query("code")),
		Name:       strings.TrimSpace(c.Query("name")),
		Teacher:    strings.TrimSpace(c.Query("teacher")),
		Department: strings.TrimSpace(c.Query("department")),
		LessonNo:   strings.TrimSpace(c.Query("lesson_no")),
		Sort:       c.DefaultQuery("sort", "code"),
		Desc:       c.Query("direction") == "desc",
		Page:       page,
		PageSize:   pageSize,
	}
	for _, bound := range []struct {
		key    string
		target **int
	}{
		{"min_sections", &filter.MinSections}, {"max_sections", &filter.MaxSections},
	} {
		if raw := c.Query(bound.key); raw != "" {
			value, parseErr := strconv.Atoi(raw)
			if parseErr != nil {
				respondBadRequest(c, "教学班数量必须是整数")
				return
			}
			*bound.target = &value
		}
	}
	if direction := c.Query("direction"); direction != "" && direction != "asc" && direction != "desc" {
		respondBadRequest(c, "排序方向无效")
		return
	}
	if filter.Semester == "" {
		respondBadRequest(c, "请先选择学期")
		return
	}
	result, err := ctrl.service.List(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *CourseAdminController) Create(c *gin.Context) {
	var offering course.ManagedOffering
	if err := c.ShouldBindJSON(&offering); err != nil {
		respondBindError(c, err)
		return
	}
	saved, err := ctrl.service.Save(c.Request.Context(), 0, offering)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func (ctrl *CourseAdminController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("offeringID"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "开课记录 ID 无效")
		return
	}
	var offering course.ManagedOffering
	if err := c.ShouldBindJSON(&offering); err != nil {
		respondBindError(c, err)
		return
	}
	saved, err := ctrl.service.Save(c.Request.Context(), id, offering)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (ctrl *CourseAdminController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("offeringID"))
	if err != nil || id <= 0 {
		respondBadRequest(c, "开课记录 ID 无效")
		return
	}
	if err := ctrl.service.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
