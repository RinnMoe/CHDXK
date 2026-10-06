package review

import (
	"context"
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

func TestGuardianCanDelete(t *testing.T) {
	ctx := context.Background()
	review := &Review{UserID: 1}

	tests := []struct {
		name string
		user *auth.User
		want bool
	}{
		{name: "nil user", user: nil, want: false},
		{name: "owner", user: &auth.User{ID: 1, Role: auth.RoleUser}, want: false},
		{name: "other user", user: &auth.User{ID: 2, Role: auth.RoleUser}, want: false},
		{name: "admin", user: &auth.User{ID: 2, Role: auth.RoleAdmin}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGuardian(tt.user, review)
			if got := g.CanDelete(ctx); got != tt.want {
				t.Fatalf("CanDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMutedAuthorCannotPublishOrEdit(t *testing.T) {
	until := time.Now().Add(time.Hour)
	u := &auth.User{ID:8,Role:auth.RoleUser,MutedUntil:&until}
	g := NewGuardian(u,&Review{UserID:8})
	if g.CanCreate(context.Background()) || g.CanUpdate(context.Background()) { t.Fatal("muted author can write") }
	u.Role = auth.RoleAdmin
	if !g.CanUpdate(context.Background()) { t.Fatal("content moderator cannot maintain existing review") }
	if g.CanCreate(context.Background()) { t.Fatal("muted moderator can publish new review") }
}
