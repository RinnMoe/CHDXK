package application

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/setting"
	"testing"
	"time"
)

func TestSelectedLocalAccountSession(t *testing.T) {
	ctx := context.Background()
	u := &auth.User{ID: 1, Role: auth.RoleSuperAdmin, AuthVersion: 3}
	repo := &auth.MockAccountRepository{User: u}
	client := &fakeResHubOIDCClient{identity: &auth.ResHubIdentity{Issuer: "https://hub.test", Subject: "user:8"}}
	service := NewAuthResolutionService(nil, nil, client, nil)
	service.SetAccounts(repo)
	for _, tc := range []struct {
		token        string
		version      int
		match, valid bool
	}{{"", 3, false, true}, {"", 2, false, false}, {"oidc-token", 3, false, false}, {"oidc-token", 3, true, true}} {
		repo.Match = tc.match
		result, err := service.ResolveSession(ctx, "", tc.token, 1, tc.version)
		if tc.valid {
			if err != nil || result.User.ID != 1 {
				t.Fatalf("local session failed: %v", err)
			}
		} else if !errors.Is(err, auth.ErrResHubSession) {
			t.Fatalf("accepted stale/unbound session: %v", err)
		}
	}
	now := time.Now()
	u.SuspendedAt = &now
	repo.Match = true
	if _, err := service.ResolveSession(ctx, "", "oidc-token", 1, 3); err == nil {
		t.Fatal("banned local account accepted")
	}
}
func TestLocalPasswordAndDelegation(t *testing.T) {
	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	target := &auth.User{ID: 8, Username: "test", Role: auth.RoleUser, LocalRole: auth.RoleUser, HasPassword: true}
	repo := &auth.MockAccountRepository{User: target, Credential: &auth.AccountCredential{User: target, PasswordHash: string(hash)}}
	s := NewLocalAccountService(repo, nil, nil, nil)
	if _, err := s.Login(ctx, "test", "wrong"); !errors.Is(err, auth.ErrLocalLogin) {
		t.Fatal("bad password accepted")
	}
	if _, err := s.Login(ctx, "test", "test-password"); err != nil {
		t.Fatal(err)
	}
	if err := s.ChangePassword(ctx, target, "wrong", "new-password"); err == nil {
		t.Fatal("changed without old password")
	}
	if err := s.ChangePassword(ctx, target, "test-password", "test-password"); err == nil {
		t.Fatal("unchanged initial password accepted")
	}
	actor := &auth.User{ID: 2, Role: auth.RoleAdmin, Permissions: []string{auth.PermissionRoles}}
	if err := s.SetRole(ctx, actor, 8, auth.RoleSuperAdmin, nil); err == nil {
		t.Fatal("created superadmin")
	}
	if err := s.SetRole(ctx, actor, 8, auth.RoleAdmin, []string{auth.PermissionPoints}); err == nil {
		t.Fatal("granted unowned permission")
	}
	root := &auth.User{ID: 1, Role: auth.RoleSuperAdmin}
	if err := s.SetRole(ctx, root, 8, auth.RoleAdmin, []string{auth.PermissionSettings}); err != nil {
		t.Fatal(err)
	}
	if target.LocalRole != auth.RoleAdmin || len(target.Permissions) != 1 {
		t.Fatal("delegation not saved")
	}
	if err := s.Moderate(ctx, root, 8, "mute", "违规", nil); err != nil || !target.IsMuted() {
		t.Fatal("mute not applied", err)
	}
	target.ID = 1
	if _, err := s.Target(ctx, actor, 1); err == nil {
		t.Fatal("owner unprotected")
	}
}

func TestRegistrationSwitchAndBoundPoints(t *testing.T) {
	ctx := context.Background()
	settingsRepo := setting.NewMockSystemRepository()
	settings := setting.NewSystemSettingsService(settingsRepo, setting.NewRegistry([]setting.Definition{{Key: RegistrationSetting, Type: setting.ValueTypeBool, DefaultValue: "false"}}))
	target := &auth.User{ID: 8, Role: auth.RoleUser}
	repo := &auth.MockAccountRepository{User: target}
	provider := &accountPointProvider{}
	service := NewLocalAccountService(repo, settings, provider, nil)
	if _, err := service.Register(ctx, "new-user", "昵称", "new-password"); err == nil {
		t.Fatal("registration closed but account created")
	}
	if _, err := settings.Save(ctx, RegistrationSetting, "true"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(ctx, "new-user", "昵称", "new-password"); err != nil {
		t.Fatal("registration enabled", err)
	}
	actor := &auth.User{ID: 1, Role: auth.RoleSuperAdmin}
	if _, err := service.AdjustPoints(ctx, actor, 8, 5, "调试发放", "request-123"); !errors.Is(err, point.ErrNoPointAccount) || provider.calls != 0 {
		t.Fatal("unbound account points attempted", err)
	}
	target.Subject = "user:8"
	target.HasPointsAccount = true
	operation, err := service.AdjustPoints(ctx, actor, 8, -5, "调试收回", "request-123")
	if err != nil || operation.Amount != -5 || provider.subject != "user:8" || provider.key != "chdxk:admin:1:request-123" {
		t.Fatal("point target/idempotency", operation, err)
	}
	if _, err := service.AdjustPoints(ctx, &auth.User{ID: 9, Role: auth.RoleAdmin}, 8, 5, "发放", "request-456"); err == nil {
		t.Fatal("unapproved admin changed points")
	}
}

type accountPointProvider struct {
	point.ExternalPointProvider
	calls        int
	subject, key string
}

func (p *accountPointProvider) Change(_ context.Context, subject string, amount int, reason, key string) (point.ExternalOperation, error) {
	p.calls++
	p.subject = subject
	p.key = key
	return point.ExternalOperation{ID: "operation-1", Amount: amount, Subject: subject, BalanceAfter: 95}, nil
}
