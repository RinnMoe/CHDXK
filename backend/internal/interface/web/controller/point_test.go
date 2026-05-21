package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/interface/web/controller"
)

func TestPointController_GetUserPointsAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		current    *auth.User
		targetID   string
		wantStatus int
	}{
		{name: "unauthorized", current: nil, targetID: "1", wantStatus: http.StatusUnauthorized},
		{name: "forbidden", current: &auth.User{ID: 1, Role: auth.RoleUser}, targetID: "2", wantStatus: http.StatusForbidden},
		{name: "self", current: &auth.User{ID: 1, Role: auth.RoleUser}, targetID: "1", wantStatus: http.StatusOK},
		{name: "admin", current: &auth.User{ID: 1, Role: auth.RoleAdmin}, targetID: "2", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &pointControllerFakeQuery{}
			ctrl := controller.NewPointController(application.NewPointQueryService(repo))
			r := gin.New()
			r.GET("/api/user/:userID/points", func(c *gin.Context) {
				if tt.current != nil {
					c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), tt.current))
				}
				ctrl.GetUserPoints(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/"+tt.targetID+"/points?page=1&page_size=20", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

type pointControllerFakeQuery struct{}

func (q *pointControllerFakeQuery) SumByUser(_ context.Context, _ int) (int, error) {
	return 0, nil
}

func (q *pointControllerFakeQuery) FindRecordsByUser(_ context.Context, filter point.RecordFilter) ([]point.Record, int64, error) {
	return []point.Record{}, 0, nil
}
