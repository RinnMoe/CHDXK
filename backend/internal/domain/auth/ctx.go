package auth

import "context"

type ctxKey string

const (
	CtxUserKey   ctxKey = "user"
	CtxApiKeyKey ctxKey = "api_key"
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

func GetApiKeyFromCtx(c context.Context) *ApiKey {
	v := c.Value(CtxApiKeyKey)
	if v == nil {
		return nil
	}
	k, ok := v.(*ApiKey)
	if !ok {
		return nil
	}
	return k
}

func WithApiKey(c context.Context, k *ApiKey) context.Context {
	return context.WithValue(c, CtxApiKeyKey, k)
}
