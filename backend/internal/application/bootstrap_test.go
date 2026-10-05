package application

import (
	"context"
	"errors"
	"testing"

	"jcourse/internal/domain/auth"
)

func TestBootstrapResolution(t *testing.T) {
	client := &fakeResHubOIDCClient{identity: &auth.ResHubIdentity{Issuer: "https://hub.test", Subject: "user:8"}}
	service := NewAuthResolutionService(nil, nil, client, nil)
	repo := &auth.MockBootstrapAccountRepository{User: &auth.User{ID: 1, Role: auth.RoleSuperAdmin, IsBootstrap: true}}
	service.SetBootstrap(repo)
	root, err := service.ResolveAccount(context.Background(), "", "", true)
	if err != nil || root.User.ID != 1 || !root.User.IsSuperAdmin() {
		t.Fatalf("local root: %+v %v", root, err)
	}
	if _, err := NewAccountQueryService().CurrentUser(context.Background(), root.User); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ResolveAccount(context.Background(), "", "oidc-token", true); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatalf("unbound SSO allowed: %v", err)
	}
	repo.Match = true
	if _, err := service.ResolveAccount(context.Background(), "", "oidc-token", true); err != nil {
		t.Fatal(err)
	}
	repo.Match = false
	if _, err := service.ResolveAccount(context.Background(), "", "oidc-token", true); !errors.Is(err, auth.ErrResHubSession) {
		t.Fatal("old binding must lose bootstrap access")
	}
}
