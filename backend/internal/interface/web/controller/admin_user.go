package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
)

type AdminUserController struct {
	query   *application.AdminUserQueryService
	command *application.AdminUserCommandService
}

type suspendUserCommand struct {
	Days int `json:"days"`
}

func NewAdminUserController(
	query *application.AdminUserQueryService,
	command *application.AdminUserCommandService,
) *AdminUserController {
	return &AdminUserController{query: query, command: command}
}

func (ctrl *AdminUserController) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	result, err := ctrl.query.FindByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, account.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctrl *AdminUserController) SuspendUser(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}
	cmd := suspendUserCommand{Days: 30}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := ctrl.command.SuspendUserForDays(c.Request.Context(), userID, cmd.Days); err != nil {
		handleAdminUserCommandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AdminUserController) ClearSuspension(c *gin.Context) {
	userID, ok := bindAdminUserID(c)
	if !ok {
		return
	}

	if err := ctrl.command.ClearSuspension(c.Request.Context(), userID); err != nil {
		handleAdminUserCommandError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func bindAdminUserID(c *gin.Context) (int, bool) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}
	return userID, true
}

func handleAdminUserCommandError(c *gin.Context, err error) {
	if errors.Is(err, account.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if errors.Is(err, application.ErrCannotSuspendAdmin) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
