package application_test

import (
	"context"
	"testing"

	"jcourse/internal/application"
	"jcourse/internal/domain/account"
	"jcourse/internal/domain/auth"
)

func TestAccountQueryService_CurrentUser(t *testing.T) {
	accountRepo := newFakeAccountRepo(map[string]*account.Account{
		"alice@example.edu": {ID: 1, Username: "alice@example.edu", Email: "alice@example.edu"},
	})
	svc := application.NewAccountQueryService(accountRepo)

	dto, err := svc.CurrentUser(context.Background(), &auth.User{ID: 1, Role: auth.RoleAdmin})
	if err != nil {
		t.Fatalf("CurrentUser: %v", err)
	}
	if dto.ID != 1 || dto.Username != "alice@example.edu" || dto.Email != "alice@example.edu" || dto.Role != auth.RoleAdmin {
		t.Fatalf("dto = %+v", dto)
	}
}
