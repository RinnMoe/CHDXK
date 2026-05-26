package verification

import (
	"testing"
	"time"
)

func TestNewCodeUsesConfig(t *testing.T) {
	now := time.Now()
	code, err := NewCode("alice@example.edu", now, Config{CodeLength: 8, CodeTTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("NewCode: %v", err)
	}
	if code.Email != "alice@example.edu" {
		t.Fatalf("email = %q, want alice@example.edu", code.Email)
	}
	if len(code.Code) != 8 {
		t.Fatalf("code length = %d, want 8", len(code.Code))
	}
	if !code.ExpiresAt.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("expires at = %v, want %v", code.ExpiresAt, now.Add(5*time.Minute))
	}
}

func TestCodeMatchesTrimmedValue(t *testing.T) {
	code := &Code{Code: "123456", ExpiresAt: time.Now().Add(time.Minute)}
	if !code.Matches(" 123456\n", time.Now()) {
		t.Fatal("expected trimmed value to match")
	}
	if code.Matches("000000", time.Now()) {
		t.Fatal("expected different value to fail")
	}
	if code.Matches("123456", time.Now().Add(2*time.Minute)) {
		t.Fatal("expected expired code to fail")
	}
}
