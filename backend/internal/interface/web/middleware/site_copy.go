package middleware

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/pkg/sitecopy"
)

func SiteCopy(query *application.SystemSettingsQueryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		values, err := query.PublicCopy(c.Request.Context())
		if err == nil {
			c.Request = c.Request.WithContext(sitecopy.WithValues(c.Request.Context(), values))
		}
		c.Next()
	}
}
