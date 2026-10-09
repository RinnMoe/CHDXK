package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"net/http"
)

type CourseSyncOperations interface {
	Start(int) (*application.CourseSyncStatus, error)
	Status(int, string) (*application.CourseSyncStatus, error)
	Sync(int, string, []string, ...bool) (*application.CourseSyncStatus, error)
	Cancel(int, string) (*application.CourseSyncStatus, error)
}
type CourseSyncController struct{ service CourseSyncOperations }

func NewCourseSyncController(service CourseSyncOperations) *CourseSyncController {
	return &CourseSyncController{service}
}
func (ctrl *CourseSyncController) Start(c *gin.Context) {
	user := auth.GetUserFromCtx(c.Request.Context())
	if user == nil {
		respondUnauthorized(c)
		return
	}
	status, err := ctrl.service.Start(user.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, status)
}
func (ctrl *CourseSyncController) Status(c *gin.Context) {
	user := auth.GetUserFromCtx(c.Request.Context())
	if user == nil {
		respondUnauthorized(c)
		return
	}
	status, err := ctrl.service.Status(user.ID, c.Param("jobID"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}
func (ctrl *CourseSyncController) Sync(c *gin.Context) {
	user := auth.GetUserFromCtx(c.Request.Context())
	if user == nil {
		respondUnauthorized(c)
		return
	}
	var cmd struct {
		Semesters []string `json:"semesters"`
		Fresh     bool     `json:"fresh"`
	}
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}
	status, err := ctrl.service.Sync(user.ID, c.Param("jobID"), cmd.Semesters, cmd.Fresh)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}
func (ctrl *CourseSyncController) Cancel(c *gin.Context) {
	user := auth.GetUserFromCtx(c.Request.Context())
	if user == nil {
		respondUnauthorized(c)
		return
	}
	status, err := ctrl.service.Cancel(user.ID, c.Param("jobID"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
}
