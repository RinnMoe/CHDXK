package canteen

import "jcourse/internal/domain/auth"

// Guardian centralises owner/admin decisions for visit reviews. Unlike the
// course domain, authors may soft-delete their own canteen reviews.
type Guardian struct {
	user   *auth.User
	review *ReviewEvent
}

func NewGuardian(user *auth.User, review *ReviewEvent) Guardian {
	return Guardian{user: user, review: review}
}

func (g Guardian) CanView() bool {
	if g.user == nil || g.review == nil {
		return false
	}
	switch g.review.Status {
	case ReviewStatusActive, ReviewStatusHidden:
		return true
	default:
		return g.user.IsAdmin() || g.user.ID == g.review.UserID
	}
}

func (g Guardian) CanUpdate() bool {
	if g.user == nil || g.review == nil {
		return false
	}
	if g.user.IsAdmin() {
		return true
	}
	return g.user.ID == g.review.UserID && !g.user.IsMuted() && g.review.Status == ReviewStatusActive
}

func (g Guardian) CanDelete() bool {
	if g.user == nil || g.review == nil {
		return false
	}
	return g.user.IsAdmin() || g.user.ID == g.review.UserID
}

func (g Guardian) CanHide() bool {
	return g.user != nil && g.user.IsAdmin()
}

func (g Guardian) CanVote() bool {
	if g.user == nil || g.review == nil {
		return false
	}
	return g.review.Status == ReviewStatusActive && g.user.ID != g.review.UserID
}

// CanCreateReview mirrors the course rule: suspended or muted accounts cannot
// publish, while administrators keep their ability to act.
func CanCreateReview(user *auth.User) bool {
	if user == nil {
		return false
	}
	if user.IsAdmin() {
		return true
	}
	return !user.IsSuspended() && !user.IsMuted()
}
