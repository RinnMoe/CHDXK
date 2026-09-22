package controller

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type AccountController struct {
	query *application.AccountQueryService
	oidc  *application.ResHubOIDCService
}

func NewAccountController(query *application.AccountQueryService, oidc *application.ResHubOIDCService) *AccountController {
	return &AccountController{query: query, oidc: oidc}
}

func (ctrl *AccountController) OIDCStart(c *gin.Context) {
	returnPath := application.SafeReturnPath(c.Query("return_to"))
	if ctrl.oidc == nil {
		ctrl.redirectOIDCError(c, returnPath, "oidc_unavailable")
		return
	}
	state, err := application.NewOIDCState()
	if err != nil {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	verifier, err := application.NewPKCEVerifier()
	if err != nil {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	nonce, err := application.NewOIDCNonce()
	if err != nil {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	if err := middleware.SetOIDCState(c, state, verifier, nonce, returnPath); err != nil {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	location, err := ctrl.oidc.StartURL(c.Request.Context(), state, verifier, nonce)
	if err != nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectOIDCError(c, returnPath, "oidc_unavailable")
		return
	}
	c.Redirect(http.StatusFound, location)
}

func (ctrl *AccountController) OIDCCallback(c *gin.Context) {
	state, verifier, nonce, returnPath, ok := middleware.TakeOIDCState(c, c.Query("state"))
	if !ok || state == "" {
		_ = middleware.ClearSession(c)
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	if ctrl.oidc == nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectOIDCError(c, returnPath, "oidc_unavailable")
		return
	}
	if remoteError := c.Query("error"); remoteError != "" {
		_ = middleware.ClearSession(c)
		ctrl.redirectOIDCError(c, returnPath, "oidc_denied")
		return
	}
	user, session, err := ctrl.oidc.Complete(c.Request.Context(), c.Query("code"), verifier, nonce)
	if err != nil || user == nil || session.AccessToken == "" || session.IDToken == "" {
		_ = middleware.ClearSession(c)
		code := "oidc_failed"
		if errors.Is(err, auth.ErrResHubUnavailable) {
			code = "oidc_unavailable"
		}
		ctrl.redirectOIDCError(c, returnPath, code)
		return
	}
	if err := middleware.SetSessionOIDC(c, session.AccessToken, session.IDToken); err != nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	c.Redirect(http.StatusFound, application.SafeReturnPath(returnPath))
}

func (ctrl *AccountController) redirectOIDCError(c *gin.Context, returnPath, code string) {
	query := url.Values{}
	query.Set("error", code)
	query.Set("redirect", application.SafeReturnPath(returnPath))
	c.Redirect(http.StatusFound, "/login?"+query.Encode())
}

func (ctrl *AccountController) Logout(c *gin.Context) {
	idToken := middleware.OIDCIDToken(c)
	var remoteErr error
	if idToken != "" && ctrl.oidc != nil {
		remoteErr = ctrl.oidc.Logout(c.Request.Context(), idToken)
	}
	clearErr := middleware.ClearSession(c)
	if clearErr != nil {
		if remoteErr != nil {
			remoteErr = errors.Join(remoteErr, clearErr)
		} else {
			remoteErr = clearErr
		}
	}
	if remoteErr != nil {
		respondError(c, remoteErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) Me(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		respondUnauthorized(c)
		return
	}
	dto, err := ctrl.query.CurrentUser(c.Request.Context(), u)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto)
}
