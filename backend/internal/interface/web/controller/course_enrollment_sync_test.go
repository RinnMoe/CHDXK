package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func TestLoadEnrollmentSyncStateTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		createdAt time.Time
		wantErr   bool
	}{
		{name: "fresh state", createdAt: time.Now().Add(-enrollmentSyncStateTTL + time.Second)},
		{name: "expired state", createdAt: time.Now().Add(-enrollmentSyncStateTTL - time.Second), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(sessions.Sessions("jcourse_session", cookie.NewStore([]byte("test-secret"))))
			r.GET("/state", func(c *gin.Context) {
				state := enrollmentSyncState{
					State:     "state",
					UserID:    1,
					Semester:  "2025-2026-1",
					CreatedAt: tt.createdAt.Unix(),
				}
				encoded, err := json.Marshal(state)
				if err != nil {
					t.Fatalf("marshal state: %v", err)
				}
				sessions.Default(c).Set(enrollmentSyncSessionKey, string(encoded))

				_, err = loadEnrollmentSyncState(c)
				if tt.wantErr && err == nil {
					t.Fatalf("loadEnrollmentSyncState error = nil, want error")
				}
				if !tt.wantErr && err != nil {
					t.Fatalf("loadEnrollmentSyncState error = %v, want nil", err)
				}
				c.Status(http.StatusNoContent)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/state", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
			}
		})
	}
}
