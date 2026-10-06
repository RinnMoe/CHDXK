package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

func TestIdentityProjectionFollowsAccountRoleRegardlessOfPermissionMode(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserIdentityRepository(db, "https://hub.example")
	snapshot := &auth.ResHubIdentity{Issuer: "https://hub.example", Subject: "user:99001", Username: "管理员", IdentityRole: "admin"}
	var id int
	for _, tc := range []struct{ identity, effective, mode, local string }{
		{"admin", "member", "member", auth.RoleAdmin},
		{"admin", "admin", "admin", auth.RoleAdmin},
		{"superadmin", "member", "member", auth.RoleAdmin},
		{"superadmin", "admin", "admin", auth.RoleAdmin},
		{"superadmin", "superadmin", "superadmin", auth.RoleAdmin},
		{"member", "admin", "admin", auth.RoleUser},
		{"member", "member", "member", auth.RoleUser},
		{"unknown", "superadmin", "superadmin", auth.RoleUser},
		{"", "superadmin", "superadmin", auth.RoleUser},
	} {
		snapshot.IdentityRole = tc.identity
		snapshot.EffectiveRole = tc.effective
		snapshot.PermissionMode = tc.mode
		user, err := repo.Upsert(context.Background(), snapshot)
		if err != nil {
			t.Fatal(err)
		}
		if id == 0 {
			id = user.ID
		}
		if user.ID != id || user.Role != tc.local {
			t.Fatalf("projection = %+v, want id %d role %s", user, id, tc.local)
		}
		var stored repository.UserEntity
		if err := db.First(&stored, id).Error; err != nil {
			t.Fatal(err)
		}
		if stored.Role != auth.RoleUser {
			t.Fatalf("stored role = %s", stored.Role)
		}
	}
	snapshot.Issuer = "https://other.example"
	if _, err := repo.Upsert(context.Background(), snapshot); err == nil {
		t.Fatal("accepted another issuer")
	}
}
