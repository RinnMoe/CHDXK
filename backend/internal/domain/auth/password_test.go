package auth

import "testing"

func TestDjangoPBKDF2SHA256PasswordHasher_VerifyDjangoHash(t *testing.T) {
	h := NewDjangoPBKDF2SHA256PasswordHasher(1)
	encoded := "pbkdf2_sha256$260000$seasalt$ftMWvEdczZQK5azuap2CQYKRjHLa1wOuMrfMiYEswYQ="

	if !h.Verify("password", encoded) {
		t.Fatal("expected Django pbkdf2_sha256 hash to verify")
	}
	if h.Verify("wrong", encoded) {
		t.Fatal("expected wrong password to fail")
	}
}

func TestDjangoPBKDF2SHA256PasswordHasher_Hash(t *testing.T) {
	h := NewDjangoPBKDF2SHA256PasswordHasher(1)
	encoded, err := h.Hash("secret")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !h.Verify("secret", encoded) {
		t.Fatal("expected generated hash to verify")
	}
}
