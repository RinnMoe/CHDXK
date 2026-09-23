package controller

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
)

const (
	maxCourseArchiveUpload  = 64 << 20
	maxCourseArchiveRequest = maxCourseArchiveUpload + (1 << 20)
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
		Semester: strings.TrimSpace(c.Query("semester")),
		Query:    strings.TrimSpace(c.Query("q")),
		Page:     page,
		PageSize: pageSize,
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

func (ctrl *CourseAdminController) ImportArchive(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCourseArchiveRequest)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		respondBadRequest(c, "请上传包含 semester.json 和 report.json 的 ZIP 文件")
		return
	}
	defer file.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".zip") {
		respondBadRequest(c, "课程数据文件必须是 ZIP")
		return
	}
	archive, err := io.ReadAll(io.LimitReader(file, maxCourseArchiveUpload+1))
	if err != nil || len(archive) > maxCourseArchiveUpload {
		respondError(c, apperr.BadRequest("课程数据 ZIP 不能超过 64 MiB"))
		return
	}
	result, err := ctrl.service.ImportArchive(c.Request.Context(), archive)
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
