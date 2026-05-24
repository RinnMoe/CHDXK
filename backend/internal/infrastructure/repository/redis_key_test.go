package repository

import (
	"testing"
	"time"

	"jcourse/internal/domain/course"
)

func TestRedisKey(t *testing.T) {
	got := redisKey("auth", "register", "code", "alice@example.edu")
	want := "jcourse:auth:register:code:alice@example.edu"
	if got != want {
		t.Fatalf("redisKey = %s, want %s", got, want)
	}
}

func TestVerificationCodeRepositoryKeys(t *testing.T) {
	repo := NewVerificationCodeRepositoryWithPrefix(nil, "reset")

	if got, want := repo.codeKey("Alice@Example.EDU"), "jcourse:auth:reset:code:alice@example.edu"; got != want {
		t.Fatalf("codeKey = %s, want %s", got, want)
	}
	if got, want := repo.cooldownKey("Alice@Example.EDU"), "jcourse:auth:reset:code_cooldown:alice@example.edu"; got != want {
		t.Fatalf("cooldownKey = %s, want %s", got, want)
	}
}

func TestLoginAttemptRepositoryKey(t *testing.T) {
	repo := NewLoginAttemptRepository(nil, time.Minute)
	got := repo.key("Alice@Example.EDU")
	want := "jcourse:auth:login_attempts:alice@example.edu"
	if got != want {
		t.Fatalf("key = %s, want %s", got, want)
	}
}

func TestCourseHotRepositoryKey(t *testing.T) {
	loc := mustShanghaiLocation(t)
	repo := NewCourseHotRepository(nil, loc)
	at := time.Date(2026, time.May, 22, 12, 0, 0, 0, loc)

	if got, want := repo.key(course.HotCoursePeriodWeek, at), "jcourse:course:hot:week:2026-20"; got != want {
		t.Fatalf("week key = %s, want %s", got, want)
	}
	if got, want := repo.key(course.HotCoursePeriodMonth, at), "jcourse:course:hot:month:2026-05"; got != want {
		t.Fatalf("month key = %s, want %s", got, want)
	}
}
