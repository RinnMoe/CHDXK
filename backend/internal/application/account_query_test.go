package application_test

import (
	"context"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
)

func TestAccountQueryService_CurrentUser(t *testing.T) {
	svc := application.NewAccountQueryService()

	dto, err := svc.CurrentUser(context.Background(), &auth.User{
		ID:           1,
		Role:         auth.RoleAdmin,
		Subject:      "user:1",
		Username:     "alice",
		Email:        "alice@example.edu",
		IdentityRole: "member",
	})
	if err != nil {
		t.Fatalf("CurrentUser: %v", err)
	}
	if dto.ID != 1 || dto.Username != "alice" || dto.Email != "alice@example.edu" || dto.Role != auth.RoleAdmin || dto.Subject != "user:1" {
		t.Fatalf("dto = %+v", dto)
	}
}
