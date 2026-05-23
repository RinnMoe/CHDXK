package review

import (
	"context"
	"testing"

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
