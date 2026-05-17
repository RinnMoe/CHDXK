package review

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

type CreatePolicy interface {
	CanCreate(ctx context.Context, u *auth.User, c *course.Course, r *Review) error
}

type Guardian struct {
	u *auth.User
	r *Review
}

func NewGuardian(u *auth.User, r *Review) Guardian {
	return Guardian{u: u, r: r}
}
func (g Guardian) CanDelete(ctx context.Context) bool {
	return g.u.ID == g.r.UserID || g.u.IsAdmin()
}
func (g Guardian) CanUpdate(ctx context.Context) bool {
	return g.u.ID == g.r.UserID || g.u.IsAdmin()
}
func (g Guardian) CanCreate(ctx context.Context) bool {
	return !g.u.IsSuspended()
}
