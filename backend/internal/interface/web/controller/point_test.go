package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
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
			ctrl := controller.NewPointController(application.NewPointQueryService(repo, &pointControllerFakeAccountRepo{}, point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}), testPointUsernameDeriver()), cmd)
			r := gin.New()
			r.GET("/api/user/:userID/point", func(c *gin.Context) {
				if tt.current != nil {
					c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), tt.current))
				}
				ctrl.GetUserPoints(c)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/user/"+tt.targetID+"/point?page=1&page_size=20", nil)
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
	ctrl := controller.NewPointController(application.NewPointQueryService(&pointControllerFakeQuery{}, &pointControllerFakeAccountRepo{}, point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}), testPointUsernameDeriver()), command)
	r := gin.New()
	r.POST("/api/point/transfer/preview", func(c *gin.Context) {
		c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 1, Role: auth.RoleUser}))
		ctrl.PreviewTransfer(c)
	})
	r.POST("/api/point/transfer", func(c *gin.Context) {
		c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 1, Role: auth.RoleUser}))
		ctrl.CreateTransfer(c)
	})

	body := `{"recipient_username":"bob@example.edu","amount":100,"fee_payer":"sender"}`
	previewReq := httptest.NewRequest(http.MethodPost, "/api/point/transfer/preview", strings.NewReader(body))
	previewReq.Header.Set("Content-Type", "application/json")
	previewW := httptest.NewRecorder()
	r.ServeHTTP(previewW, previewReq)
	if previewW.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200", previewW.Code)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/point/transfer", strings.NewReader(body))
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
		&pointControllerFakeAccountRepo{accountsByID: map[int]*account.Account{
			1: {ID: 1, Username: "alice@example.edu", Email: "alice@example.edu"},
		}, accountsByUsername: map[string]*account.Account{
			"bob@example.edu": {ID: 2, Username: "bob@example.edu", Email: "bob@example.edu"},
		}},
		&pointControllerFakeTransferRepo{},
		point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}),
	)
}

func testPointUsernameDeriver() account.UsernameDeriver {
	return account.NewBLAKE2bUsernameDeriver(account.UsernameDeriverConfig{Salt: "SALT"})
}

type pointControllerFakeAccountRepo struct {
	accountsByID       map[int]*account.Account
	accountsByUsername map[string]*account.Account
	accountsByEmail    map[string]*account.Account
}

func (r *pointControllerFakeAccountRepo) Create(_ context.Context, _ *account.Account) error {
	return nil
}
func (r *pointControllerFakeAccountRepo) Update(_ context.Context, _ *account.Account) error {
	return nil
}
func (r *pointControllerFakeAccountRepo) TouchLastSeen(_ context.Context, _ int, _ time.Time) error {
	return nil
}
func (r *pointControllerFakeAccountRepo) FindByID(_ context.Context, id int) (*account.Account, error) {
	return r.accountsByID[id], nil
}
func (r *pointControllerFakeAccountRepo) FindByUsername(_ context.Context, username string) (*account.Account, error) {
	return r.accountsByUsername[username], nil
}
func (r *pointControllerFakeAccountRepo) FindByEmail(_ context.Context, email string) (*account.Account, error) {
	return r.accountsByEmail[email], nil
}

type pointControllerFakeTransferRepo struct{}

func (r *pointControllerFakeTransferRepo) CreateTransfer(_ context.Context, t *point.Transfer, _ point.Record, _ point.Record) error {
	t.ID = 1
	return nil
}

func TestPointController_GetPointsByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	email := "alice@example.edu"
	username, err := testPointUsernameDeriver().UsernameFromEmail(email)
	if err != nil {
		t.Fatalf("UsernameFromEmail: %v", err)
	}
	fakeQuery := &pointControllerFakeQueryWithEmail{
		balances: map[int]int{1: 42},
	}
	fakeAccountRepo := &pointControllerFakeAccountRepo{
		accountsByUsername: map[string]*account.Account{
			username: {ID: 1, Username: username},
		},
	}
	ctrl := controller.NewPointController(
		application.NewPointQueryService(fakeQuery, fakeAccountRepo, point.NewTransferService(point.TransferFeeConfig{RateBps: 250, MinFee: 1}), testPointUsernameDeriver()),
		newPointControllerCommand(),
	)

	r := gin.New()
	r.GET("/api/ext/point", ctrl.GetPointsByEmail)

	t.Run("missing email", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ext/point", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ext/point?email=unknown@example.edu", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/ext/point?email="+email, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp["email"] != email {
			t.Fatalf("email = %v, want %v", resp["email"], email)
		}
		total, ok := resp["total"].(float64)
		if !ok || total != 42 {
			t.Fatalf("total = %v, want 42", resp["total"])
		}
	})
}

type pointControllerFakeQueryWithEmail struct {
	balances map[int]int
}

func (q *pointControllerFakeQueryWithEmail) SumByUser(_ context.Context, userID int) (int, error) {
	if v, ok := q.balances[userID]; ok {
		return v, nil
	}
	return 0, nil
}

func (q *pointControllerFakeQueryWithEmail) FindRecordsByUser(_ context.Context, _ point.RecordFilter) ([]point.Record, int64, error) {
	return nil, 0, nil
}
