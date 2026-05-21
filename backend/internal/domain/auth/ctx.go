package auth

import "context"

type ctxKey string

const (
	CtxUserKey ctxKey = "user"
)

func GetUserFromCtx(c context.Context) *User {
	v := c.Value(CtxUserKey)
	if v == nil {
		return nil
	}
	u, ok := v.(*User)
	if !ok {
		return nil
	}
	return u
}

func WithUser(c context.Context, u *User) context.Context {
	return context.WithValue(c, CtxUserKey, u)
}
