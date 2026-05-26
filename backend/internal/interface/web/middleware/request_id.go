package middleware

import (
	"github.com/gin-gonic/gin"

	"jcourse/pkg/requestid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get(requestid.GinKey); exists {
			c.Next()
			return
		}

		id := c.GetHeader(requestid.HeaderRequestID)
		if id == "" {
			id = requestid.New()
			c.Request.Header.Set(requestid.HeaderRequestID, id)
		}

		ctx := requestid.WithContext(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)
		c.Header(requestid.HeaderRequestID, id)
		c.Set(requestid.GinKey, id)
		c.Next()
	}
}
