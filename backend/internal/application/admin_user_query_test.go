package application

import (
	"context"
	"testing"

	"jcourse/internal/domain/account/identity"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
)

type adminUserQueryUsernameDeriver struct {
	username string
}

func (d adminUserQueryUsernameDeriver) UsernameFromEmail(string) (string, error) {
	return d.username, nil
}

type adminUserQueryReviewQuery struct {
	reviews map[int]review.ReviewView
}

func (q adminUserQueryReviewQuery) FindBy(context.Context, review.ReviewFilter) ([]review.ReviewView, int64, error) {
	return nil, 0, nil
}

func (q adminUserQueryReviewQuery) GetByID(_ context.Context, reviewID int) (*review.ReviewView, error) {
	rv, ok := q.reviews[reviewID]
	if !ok {
		return nil, nil
	}
	return &rv, nil
}

func (q adminUserQueryReviewQuery) GetCourseFilters(context.Context, int) (*review.ReviewFilters, error) {
	return nil, nil
}

func (q adminUserQueryReviewQuery) GetCourseTrend(context.Context, int) ([]review.ReviewTrendItem, error) {
	return nil, nil
}

func (q adminUserQueryReviewQuery) FindRevisions(context.Context, int) ([]review.RevisionView, error) {
	return nil, nil
}

func TestAdminUserQueryServiceFindUser(t *testing.T) {
	ctx := context.Background()
	accounts := identity.NewMockRepository(nil)
	accounts.PutAccount("", &identity.Account{ID: 1, Username: "alice", Email: "alice@example.edu"})
	accounts.PutAccount("", &identity.Account{ID: 2, Username: "bob", Email: "bob@example.edu"})
	accounts.PutAccount("", &identity.Account{ID: 3, Username: "derived"})
	users := auth.NewMockUserRepository(map[int]*auth.User{
		1: {ID: 1, Role: auth.RoleUser},
		2: {ID: 2, Role: auth.RoleUser},
		3: {ID: 3, Role: auth.RoleUser},
	})
	reviews := adminUserQueryReviewQuery{reviews: map[int]review.ReviewView{
		10: {ID: 10, UserID: 2},
	}}
	service := NewAdminUserQueryService(
		accounts,
		users,
		reviews,
		adminUserQueryUsernameDeriver{username: "derived"},
	)

	cases := []struct {
		name         string
		lookup       AdminUserLookup
		wantID       int
		wantEmail    string
		wantUsername string
	}{
		{
			name:         "by email",
			lookup:       AdminUserLookup{Email: " Alice@Example.EDU "},
			wantID:       1,
			wantEmail:    "alice@example.edu",
			wantUsername: "alice",
		},
		{
			name:         "by username",
			lookup:       AdminUserLookup{Username: " bob "},
			wantID:       2,
			wantEmail:    "bob@example.edu",
			wantUsername: "bob",
		},
		{
			name:         "by review id",
			lookup:       AdminUserLookup{ReviewID: 10},
			wantID:       2,
			wantEmail:    "bob@example.edu",
			wantUsername: "bob",
		},
		{
			name:         "email fallback to derived username",
			lookup:       AdminUserLookup{Email: "charlie@example.edu"},
			wantID:       3,
			wantEmail:    "charlie@example.edu",
			wantUsername: "derived",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := service.FindUser(ctx, tc.lookup)
			if err != nil {
				t.Fatalf("FindUser: %v", err)
			}
			if got.ID != tc.wantID || got.Email != tc.wantEmail || got.Username != tc.wantUsername {
				t.Fatalf("FindUser = id %d email %q username %q, want id %d email %q username %q", got.ID, got.Email, got.Username, tc.wantID, tc.wantEmail, tc.wantUsername)
			}
		})
	}
}
