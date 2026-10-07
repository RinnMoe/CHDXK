package controller

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

var nativeStatePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)
var nativeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)
var nativeCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{32,200}$`)

const nativeSelectionTTL = 5 * time.Minute

func nativeAuthError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "internal_error", "暂时无法登录，请稍后重试"
	switch {
	case errors.Is(err, auth.ErrResHubUnavailable):
		status, code, message = 503, "provider_unavailable", "资料站暂时无法连接"
	case errors.Is(err, auth.ErrResHubSession), errors.Is(err, auth.ErrResHubState):
		status, code, message = 401, "invalid_grant", "授权已失效，请重新登录"
	case errors.Is(err, auth.ErrUserSuspended):
		status, code, message = 403, "account_suspended", "该账号已被封禁"
	}
	c.JSON(status, gin.H{"code": code, "error": message})
}

func (ctrl *AccountController) NativeExchange(c *gin.Context) {
	var body struct {
		Code     string `json:"code"`
		Verifier string `json:"code_verifier"`
		State    string `json:"state"`
	}
	if c.ShouldBindJSON(&body) != nil || !nativeCodePattern.MatchString(body.Code) || !nativeVerifierPattern.MatchString(body.Verifier) || !nativeStatePattern.MatchString(body.State) {
		c.JSON(400, gin.H{"code": "invalid_request", "error": "授权参数无效"})
		return
	}
	identity, session, err := ctrl.oidc.CompleteNativeIdentity(c.Request.Context(), body.Code, body.Verifier, body.State)
	if err != nil {
		nativeAuthError(c, err)
		return
	}
	if identity == nil || session.AccessToken == "" || session.IDToken == "" {
		nativeAuthError(c, auth.ErrResHubSession)
		return
	}
	choices, err := ctrl.accounts.Repo.ResHubAccounts(c.Request.Context(), identity, true)
	if err != nil {
		nativeAuthError(c, err)
		return
	}
	if len(choices) == 0 {
		nativeAuthError(c, auth.ErrResHubSession)
		return
	}
	if len(choices) == 1 {
		if err := auth.EnsureUserActive(c.Request.Context(), choices[0]); err != nil {
			nativeAuthError(c, err)
			return
		}
	}
	if err := middleware.SetSessionOIDC(c, session.AccessToken, session.IDToken); err != nil {
		nativeAuthError(c, err)
		return
	}
	if len(choices) > 1 {
		if err := middleware.SetNativePending(c, nativeSelectionTTL); err != nil {
			_ = middleware.ClearSession(c)
			nativeAuthError(c, err)
			return
		}
		c.JSON(200, gin.H{"status": "account_selection_required", "expires_in": int(nativeSelectionTTL.Seconds())})
		return
	}
	if err := middleware.SetAccountSession(c, choices[0].ID, choices[0].AuthVersion, false); err != nil {
		_ = middleware.ClearSession(c)
		nativeAuthError(c, err)
		return
	}
	c.JSON(200, gin.H{"status": "authenticated", "account": application.NewAccountDTO(choices[0])})
}
