package canteen

import (
	"testing"
	"time"

	"jcourse/internal/domain/auth"
)

func testUser(id int, role string) *auth.User {
	return &auth.User{ID: id, Role: role}
}

func TestGuardianViewAndUpdate(t *testing.T) {
	review := &ReviewEvent{ID: 10, UserID: 2, Status: ReviewStatusActive}
	author := testUser(2, auth.RoleUser)
	other := testUser(3, auth.RoleUser)
	admin := testUser(1, auth.RoleAdmin)

	if !NewGuardian(author, review).CanUpdate() {
		t.Fatal("author must update")
	}
	if NewGuardian(other, review).CanUpdate() {
		t.Fatal("other user must not update")
	}
	if !NewGuardian(admin, review).CanUpdate() {
		t.Fatal("admin must update")
	}
	if !NewGuardian(other, review).CanView() {
		t.Fatal("active review must be viewable")
	}
	if !NewGuardian(other, review).CanVote() {
		t.Fatal("other user must vote")
	}
	if NewGuardian(author, review).CanVote() {
		t.Fatal("author must not vote on own review")
	}
}

func TestGuardianMutedAuthor(t *testing.T) {
	mutedUntil := time.Now().Add(time.Hour)
	author := testUser(2, auth.RoleUser)
	author.MutedUntil = &mutedUntil
	review := &ReviewEvent{ID: 10, UserID: 2, Status: ReviewStatusActive}

	if NewGuardian(author, review).CanUpdate() {
		t.Fatal("muted author must not update")
	}
	if !NewGuardian(author, review).CanDelete() {
		t.Fatal("muted author may still delete own review")
	}
}

func TestGuardianDeletedVisibility(t *testing.T) {
	review := &ReviewEvent{ID: 10, UserID: 2, Status: ReviewStatusDeleted}
	author := testUser(2, auth.RoleUser)
	other := testUser(3, auth.RoleUser)
	admin := testUser(1, auth.RoleAdmin)

	if NewGuardian(other, review).CanView() {
		t.Fatal("deleted review must not be visible to other users")
	}
	if !NewGuardian(author, review).CanView() {
		t.Fatal("author must still see own deleted review")
	}
	if !NewGuardian(admin, review).CanView() {
		t.Fatal("admin must see deleted review")
	}
}

func TestGuardianHide(t *testing.T) {
	review := &ReviewEvent{ID: 10, UserID: 2, Status: ReviewStatusActive}
	if !NewGuardian(testUser(1, auth.RoleAdmin), review).CanHide() {
		t.Fatal("admin must hide")
	}
	if NewGuardian(testUser(2, auth.RoleUser), review).CanHide() {
		t.Fatal("author must not hide")
	}
	review.Status = ReviewStatusHidden
	if NewGuardian(testUser(3, auth.RoleUser), review).CanVote() {
		t.Fatal("hidden review must not accept votes")
	}
}

func TestCanCreateReview(t *testing.T) {
	if CanCreateReview(nil) {
		t.Fatal("nil user must not create")
	}
	if !CanCreateReview(testUser(2, auth.RoleUser)) {
		t.Fatal("normal user must create")
	}
	if !CanCreateReview(testUser(1, auth.RoleAdmin)) {
		t.Fatal("admin must create")
	}

	mutedUntil := time.Now().Add(time.Hour)
	muted := testUser(2, auth.RoleUser)
	muted.MutedUntil = &mutedUntil
	if CanCreateReview(muted) {
		t.Fatal("muted user must not create")
	}

	suspended := testUser(2, auth.RoleUser)
	suspended.SuspendTill = &mutedUntil
	if CanCreateReview(suspended) {
		t.Fatal("suspended user must not create")
	}
}
