package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/account/identity"
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
			ctrl := controller.NewPointController(application.NewPointQueryService(repo, &pointControllerFakeAccountRepo{}, testPointUsernameDeriver()))
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

func testPointUsernameDeriver() identity.UsernameDeriver {
	return identity.NewBLAKE2bUsernameDeriver(identity.UsernameDeriverConfig{Salt: "SALT"})
}

type pointControllerFakeAccountRepo struct {
	accountsByID       map[int]*identity.Account
	accountsByUsername map[string]*identity.Account
	accountsByEmail    map[string]*identity.Account
}

func (r *pointControllerFakeAccountRepo) Create(_ context.Context, _ *identity.Account) error {
	return nil
}
func (r *pointControllerFakeAccountRepo) Update(_ context.Context, _ *identity.Account) error {
	return nil
}
func (r *pointControllerFakeAccountRepo) FindByID(_ context.Context, id int) (*identity.Account, error) {
	return r.accountsByID[id], nil
}
func (r *pointControllerFakeAccountRepo) FindByUsername(_ context.Context, username string) (*identity.Account, error) {
	return r.accountsByUsername[username], nil
}
func (r *pointControllerFakeAccountRepo) FindByEmail(_ context.Context, email string) (*identity.Account, error) {
	return r.accountsByEmail[email], nil
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
		accountsByUsername: map[string]*identity.Account{
			username: {ID: 1, Username: username},
		},
	}
	ctrl := controller.NewPointController(
		application.NewPointQueryService(fakeQuery, fakeAccountRepo, testPointUsernameDeriver()),
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

		var resp map[string]any
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
