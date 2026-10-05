package repository_test

import (
	"context"
	"testing"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

func TestBootstrapMagicLinkAndBinding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := repository.NewBootstrapAccountRepository(db)
	token, err := repo.Ensure(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.Verify(ctx, token); err != nil || !ok {
		t.Fatal("initial link failed", err)
	}
	if token2, err := repo.Ensure(ctx, false); err != nil || token2 != "" {
		t.Fatal("restart must preserve original link", err)
	}
	if ok, _ := repo.Verify(ctx, "invalid"); ok {
		t.Fatal("bad link allowed")
	}
	root, err := repo.Current(ctx)
	if err != nil || root.ID != 1 || root.HasPointsAccount || !root.IsBootstrap {
		t.Fatalf("unbound root: %+v %v", root, err)
	}
	identity := &auth.ResHubIdentity{Issuer: "https://hub.test", Subject: "user:8", Username: "昵称", HasPointsAccount: true}
	if err := repo.Bind(ctx, identity); err != nil {
		t.Fatal(err)
	}
	root, err = repo.Current(ctx)
	if err != nil || !root.HasPointsAccount || root.Subject != "user:8" {
		t.Fatalf("binding: %+v %v", root, err)
	}
	if ok, _ := repo.Verify(ctx, token); !ok {
		t.Fatal("binding revoked magic link")
	}
	if ok, _ := repo.Matches(ctx, "https://other.test", "user:8"); ok {
		t.Fatal("issuer mismatch accepted")
	}
	projection := repository.NewUserIdentityRepository(db, "https://hub.test")
	if subject, err := projection.SubjectByUserID(ctx, 1); err != nil || subject != "user:8" {
		t.Fatal(subject, err)
	}
	rotated, err := repo.Ensure(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := repo.Verify(ctx, token); ok {
		t.Fatal("rotation kept old secret")
	}
	if ok, _ := repo.Verify(ctx, rotated); !ok {
		t.Fatal("rotated link failed")
	}
}
