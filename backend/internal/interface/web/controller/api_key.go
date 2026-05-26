package controller

import (
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
		respondUnauthorized(c)
		return
	}

	keys, err := ctrl.query.ListMyApiKeys(c.Request.Context(), u.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (ctrl *ApiKeyController) ListSystemApiKeys(c *gin.Context) {
	keys, err := ctrl.query.ListSystemApiKeys(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (ctrl *ApiKeyController) CreateMyApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var cmd application.CreateApiKeyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	key, err := ctrl.command.CreateMyApiKey(c.Request.Context(), u.ID, cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (ctrl *ApiKeyController) CreateSystemApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	var cmd application.CreateApiKeyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		respondBindError(c, err)
		return
	}

	key, err := ctrl.command.CreateSystemApiKey(c.Request.Context(), u, cmd)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (ctrl *ApiKeyController) DeleteMyApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	id, err := strconv.ParseInt(c.Param("apiKeyID"), 10, 64)
	if err != nil || id <= 0 {
		respondBadRequest(c, "API Key ID 无效")
		return
	}

	if err := ctrl.command.DeleteMyApiKey(c.Request.Context(), u.ID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ctrl *ApiKeyController) DeleteSystemApiKey(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}

	id, err := strconv.ParseInt(c.Param("apiKeyID"), 10, 64)
	if err != nil || id <= 0 {
		respondBadRequest(c, "API Key ID 无效")
		return
	}

	if err := ctrl.command.DeleteSystemApiKey(c.Request.Context(), u, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
