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
    email_whitelist:
      - "@example.edu"
  verification:
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

	if conf.Auth.Verification.CodeInterval != 2*time.Minute {
		t.Fatalf("auth override code interval = %s, want %s", conf.Auth.Verification.CodeInterval, 2*time.Minute)
	}
	if conf.Auth.Verification.CodeLength != 8 {
		t.Fatalf("auth override code length = %d, want 8", conf.Auth.Verification.CodeLength)
	}
	if len(conf.Server.Cors.AllowedOrigins) != 2 || conf.Server.Cors.AllowedOrigins[0] != "http://localhost:5173" || conf.Server.Cors.AllowedOrigins[1] != "http://127.0.0.1:5173" {
		t.Fatalf("server default cors origins = %#v", conf.Server.Cors.AllowedOrigins)
	}
	if conf.Server.Debug {
		t.Fatal("server default debug = true, want false")
	}
	if conf.Auth.Login.Lockout != 15*time.Minute {
		t.Fatalf("auth default lockout = %s, want %s", conf.Auth.Login.Lockout, 15*time.Minute)
	}
	if conf.Auth.Access.FlushCron != "*/5 * * * *" {
		t.Fatalf("auth access default flush cron = %q, want */5 * * * *", conf.Auth.Access.FlushCron)
	}
	if conf.Auth.Access.FlushBatchSize != 1000 {
		t.Fatalf("auth access default batch size = %d, want 1000", conf.Auth.Access.FlushBatchSize)
	}
	if conf.Session.MaxAge != 2592000 {
		t.Fatalf("session default max age = %d, want 2592000", conf.Session.MaxAge)
	}
	if conf.Session.Secure {
		t.Fatal("session default secure = true, want false")
	}
	if conf.Course.RatingScore.PriorCount != 5 {
		t.Fatalf("course rating score prior count = %d, want 5", conf.Course.RatingScore.PriorCount)
	}
	if conf.Course.RatingScore.RefreshCron != "0 5 * * *" {
		t.Fatalf("course rating score refresh cron = %q, want 0 5 * * *", conf.Course.RatingScore.RefreshCron)
	}
	if !conf.Course.RatingScore.SchedulerEnabled {
		t.Fatal("course rating score scheduler enabled = false, want true")
	}
	if conf.Review.FrequencyPolicy.Window != 2*time.Hour {
		t.Fatalf("review override window = %s, want %s", conf.Review.FrequencyPolicy.Window, 2*time.Hour)
	}
	if conf.Review.Command.Vote.MaxDailyVotes != 100 {
		t.Fatalf("review override max daily votes = %d, want 100", conf.Review.Command.Vote.MaxDailyVotes)
	}
}

func TestLoadOverridesCORSAllowedOrigins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`
server:
  cors:
    allowed_origins:
      - "https://jcourse.example.edu"
      - "https://admin.example.edu"
session:
  secret: "replace-with-at-least-32-random-characters"
auth:
  username_deriver:
    salt: "SALT"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := []string{"https://jcourse.example.edu", "https://admin.example.edu"}
	if len(conf.Server.Cors.AllowedOrigins) != len(want) {
		t.Fatalf("cors origins len = %d, want %d (%#v)", len(conf.Server.Cors.AllowedOrigins), len(want), conf.Server.Cors.AllowedOrigins)
	}
	for i := range want {
		if conf.Server.Cors.AllowedOrigins[i] != want[i] {
			t.Fatalf("cors origin[%d] = %q, want %q", i, conf.Server.Cors.AllowedOrigins[i], want[i])
		}
	}
}

func TestLoadServerOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`
server:
  port: 9090
  debug: true
session:
  secret: "replace-with-at-least-32-random-characters"
auth:
  username_deriver:
    salt: "SALT"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	conf, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !conf.Server.Debug {
		t.Fatal("server debug = false, want true")
	}
}
