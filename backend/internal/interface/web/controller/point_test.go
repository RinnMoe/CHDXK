package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
			cmd := newPointControllerCommand()
			ctrl := controller.NewPointController(application.NewPointQueryService(repo), cmd)
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

func TestPointController_PreviewAndCreateTransfer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	command := newPointControllerCommand()
	ctrl := controller.NewPointController(application.NewPointQueryService(&pointControllerFakeQuery{}), command)
	r := gin.New()
	r.POST("/api/point/transfers/preview", func(c *gin.Context) {
		c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 1, Role: auth.RoleUser}))
		ctrl.PreviewTransfer(c)
	})
	r.POST("/api/point/transfers", func(c *gin.Context) {
		c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 1, Role: auth.RoleUser}))
		ctrl.CreateTransfer(c)
	})

	body := `{"recipient_username":"bob@example.edu","amount":100,"fee_payer":"sender"}`
	previewReq := httptest.NewRequest(http.MethodPost, "/api/point/transfers/preview", strings.NewReader(body))
	previewReq.Header.Set("Content-Type", "application/json")
	previewW := httptest.NewRecorder()
	r.ServeHTTP(previewW, previewReq)
	if previewW.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200", previewW.Code)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/point/transfers", strings.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", createW.Code)
	}
}

type pointControllerFakeQuery struct{}

func (q *pointControllerFakeQuery) SumByUser(_ context.Context, _ int) (int, error) {
	return 0, nil
}

func (q *pointControllerFakeQuery) FindRecordsByUser(_ context.Context, filter point.RecordFilter) ([]point.Record, int64, error) {
	return []point.Record{}, 0, nil
}

func newPointControllerCommand() *application.PointCommandService {
	return application.NewPointCommandService(
		&pointControllerFakeUserRepo{usersByUsername: map[string]*auth.User{
			"bob@example.edu": &auth.User{ID: 2, Username: "bob@example.edu", Email: "bob@example.edu", Role: auth.RoleUser},
		}},
		&pointControllerFakeTransferRepo{},
		application.PointTransferFeeConfig{RateBps: 250, MinFee: 1},
	)
}

type pointControllerFakeUserRepo struct {
	usersByUsername map[string]*auth.User
}

func (r *pointControllerFakeUserRepo) Create(_ context.Context, _ *auth.User) error { return nil }
func (r *pointControllerFakeUserRepo) Update(_ context.Context, _ *auth.User) error { return nil }
func (r *pointControllerFakeUserRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error {
	return nil
}
func (r *pointControllerFakeUserRepo) FindByID(_ context.Context, _ int) (*auth.User, error) {
	return nil, nil
}
func (r *pointControllerFakeUserRepo) FindByUsername(_ context.Context, username string) (*auth.User, error) {
	return r.usersByUsername[username], nil
}
func (r *pointControllerFakeUserRepo) FindByEmail(_ context.Context, _ string) (*auth.User, error) {
	return nil, nil
}

type pointControllerFakeTransferRepo struct{}

func (r *pointControllerFakeTransferRepo) CreateTransfer(_ context.Context, t *point.Transfer, _ point.Record, _ point.Record) error {
	t.ID = 1
	return nil
}
