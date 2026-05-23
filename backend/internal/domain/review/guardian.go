package review

import (
	"context"

	"jcourse/internal/domain/auth"
)

type ViewGuardian struct {
	u *auth.User
	v *ReviewView
}

func NewViewGuardian(u *auth.User, v *ReviewView) ViewGuardian {
	return ViewGuardian{u: u, v: v}
}

func (g ViewGuardian) CanViewPrivate() bool {
	return g.u != nil && (g.u.ID == g.v.UserID || g.u.IsAdmin())
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
func (g Guardian) CanUpdateModeratorRemark(ctx context.Context) bool {
	return g.u != nil && g.u.IsAdmin()
}
func (g Guardian) CanCreate(ctx context.Context) bool {
	return !g.u.IsSuspended()
}
