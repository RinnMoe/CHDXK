package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/point"
	"jcourse/internal/interface/web/controller"
)

func TestPointController_GetUserPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		targetID   string
		wantStatus int
	}{
		{name: "invalid user id", targetID: "bad", wantStatus: http.StatusBadRequest},
		{name: "valid user id", targetID: "1", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &pointControllerFakeQuery{}
			ctrl := controller.NewPointController(application.NewPointQueryService(repo))
			r := gin.New()
			r.GET("/api/user/:userID/point", ctrl.GetUserPoints)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/"+tt.targetID+"/point?page=1&page_size=20", nil)
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
