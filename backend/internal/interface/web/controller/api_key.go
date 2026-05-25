package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

type ApiKeyController struct {
	query   *application.ApiKeyQueryService
	command *application.ApiKeyCommandService
}

func NewApiKeyController(query *application.ApiKeyQueryService, command *application.ApiKeyCommandService) *ApiKeyController {
	return &ApiKeyController{query: query, command: command}
}

func (ctrl *ApiKeyController) ListMyApiKeys(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	keys, err := ctrl.query.ListMyApiKeys(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (ctrl *ApiKeyController) ListSystemApiKeys(c *gin.Context) {
	keys, err := ctrl.query.ListSystemApiKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (ctrl *ApiKeyController) CreateMyApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var cmd application.CreateApiKeyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := ctrl.command.CreateMyApiKey(c.Request.Context(), u.ID, cmd)
	if err != nil {
		if errors.Is(err, auth.ErrApiKeyNameRequired) || errors.Is(err, auth.ErrApiKeyLimitExceeded) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (ctrl *ApiKeyController) CreateSystemApiKey(c *gin.Context) {
	var cmd application.CreateApiKeyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := ctrl.command.CreateSystemApiKey(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, auth.ErrApiKeyNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (ctrl *ApiKeyController) DeleteMyApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.Atoi(c.Param("apiKeyID"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api key id"})
		return
	}

	if err := ctrl.command.DeleteMyApiKey(c.Request.Context(), u.ID, id); err != nil {
		if errors.Is(err, auth.ErrApiKeyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (ctrl *ApiKeyController) DeleteSystemApiKey(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("apiKeyID"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid api key id"})
		return
	}

	if err := ctrl.command.DeleteSystemApiKey(c.Request.Context(), id); err != nil {
		if errors.Is(err, auth.ErrApiKeyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
