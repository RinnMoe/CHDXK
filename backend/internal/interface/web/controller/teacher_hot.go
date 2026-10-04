package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"net/http"
	"strconv"
)

func TeacherHotHandler(service *application.TeacherHotQueryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if err != nil || limit < 1 || limit > 100 {
			respondBadRequest(c, "数量应在 1 到 100 之间")
			return
		}
		result, err := service.List(c.Request.Context(), c.DefaultQuery("period", "week"), limit)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
