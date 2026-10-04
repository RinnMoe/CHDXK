package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

func TestIdentityProjectionFollowsCurrentResHubPermissions(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserIdentityRepository(db, "https://hub.example")
	snapshot := &auth.ResHubIdentity{Issuer: "https://hub.example", Subject: "user:99001", Username: "管理员", IdentityRole: "admin"}
	var id int
	for _, tc := range []struct{ remote, local string }{
		{"admin", auth.RoleAdmin}, {"member", auth.RoleUser}, {"superadmin", auth.RoleSuperAdmin}, {"unknown", auth.RoleUser},
	} {
		snapshot.EffectiveRole = tc.remote
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
		if stored.Role != tc.local {
			t.Fatalf("stored role = %s", stored.Role)
		}
	}
	snapshot.Issuer = "https://other.example"
	if _, err := repo.Upsert(context.Background(), snapshot); err == nil {
		t.Fatal("accepted another issuer")
	}
}
