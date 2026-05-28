package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindReviewListFilter_DefaultOrderBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("keeps order_by empty when q is present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/reviews?q=data+structure", nil)

		f, err := bindReviewListFilter(c)
		if err != nil {
			t.Fatalf("bindReviewListFilter: %v", err)
		}
		if f.OrderBy != "" {
			t.Fatalf("OrderBy = %q, want empty", f.OrderBy)
		}
	})

	t.Run("defaults order_by to updated_at when q is absent", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/reviews", nil)

		f, err := bindReviewListFilter(c)
		if err != nil {
			t.Fatalf("bindReviewListFilter: %v", err)
		}
		if f.OrderBy != "updated_at" {
			t.Fatalf("OrderBy = %q, want updated_at", f.OrderBy)
		}
	})
}
