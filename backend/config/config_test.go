package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMergesSectionDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`
session:
  secret: "replace-with-at-least-32-random-characters"
auth:
  registration:
    code_interval: "2m"
    code_length: 8
  username_deriver:
    salt: "SALT"
review:
  command:
    vote:
      max_daily_votes: 100
  frequency_policy:
    window: "2h"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if conf.Auth.Registration.CodeInterval != 2*time.Minute {
		t.Fatalf("auth override code interval = %s, want %s", conf.Auth.Registration.CodeInterval, 2*time.Minute)
	}
	if conf.Auth.Registration.CodeLength != 8 {
		t.Fatalf("auth override code length = %d, want 8", conf.Auth.Registration.CodeLength)
	}
	if conf.Auth.Login.Lockout != 15*time.Minute {
		t.Fatalf("auth default lockout = %s, want %s", conf.Auth.Login.Lockout, 15*time.Minute)
	}
	if conf.Session.MaxAge != 2592000 {
		t.Fatalf("session default max age = %d, want 2592000", conf.Session.MaxAge)
	}
	if conf.Session.Secure {
		t.Fatal("session default secure = true, want false")
	}
	if conf.Review.FrequencyPolicy.Window != 2*time.Hour {
		t.Fatalf("review override window = %s, want %s", conf.Review.FrequencyPolicy.Window, 2*time.Hour)
	}
	if conf.Review.Command.Vote.MaxDailyVotes != 100 {
		t.Fatalf("review override max daily votes = %d, want 100", conf.Review.Command.Vote.MaxDailyVotes)
	}
}
