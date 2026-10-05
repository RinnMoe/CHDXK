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
	bootstrap  auth.BootstrapAccountRepository
	client     auth.ResHubOIDCClient
	resolution *application.AuthResolutionService
	query      *application.AccountQueryService
	oidc       *application.ResHubOIDCService
}

func NewAccountController(query *application.AccountQueryService, oidc *application.ResHubOIDCService) *AccountController {
	return &AccountController{query: query, oidc: oidc}
}

func (ctrl *AccountController) OIDCStart(c *gin.Context) {
	returnPath := application.SafeReturnPath(c.Query("return_to"))
	binding := c.Query("purpose") == "bind"
	if binding {
		if !middleware.BootstrapSession(c) || ctrl.bootstrap == nil {
			respondUnauthorized(c)
			return
		}
		resolved, err := ctrl.resolution.ResolveAccount(c.Request.Context(), "", middleware.OIDCAccessToken(c), true)
		if err != nil || resolved.User == nil || !resolved.User.IsBootstrap {
			respondUnauthorized(c)
			return
		}
	}
	if err := middleware.SetOIDCBinding(c, binding); err != nil {
		respondError(c, err)
		return
	}
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
	binding := middleware.TakeOIDCBinding(c)
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
	if ctrl.bootstrap != nil && ctrl.client != nil {
		identity, err := ctrl.client.UserInfo(c.Request.Context(), session.AccessToken)
		if err != nil {
			_ = middleware.ClearSession(c)
			ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
			return
		}
		if binding {
			if err := ctrl.bootstrap.Bind(c.Request.Context(), identity); err != nil {
				respondError(c, err)
				return
			}
			if err := middleware.SetBootstrapSession(c, true, false); err != nil {
				respondError(c, err)
				return
			}
			c.Redirect(http.StatusFound, "/settings")
			return
		}
		match, err := ctrl.bootstrap.Matches(c.Request.Context(), identity.Issuer, identity.Subject)
		if err != nil {
			respondError(c, err)
			return
		}
		if match {
			q := url.Values{"redirect": []string{application.SafeReturnPath(returnPath)}}
			c.Redirect(http.StatusFound, "/login/accounts?"+q.Encode())
			return
		}
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

func (ctrl *AccountController) SetBootstrap(repo auth.BootstrapAccountRepository, client auth.ResHubOIDCClient, resolution *application.AuthResolutionService) {
	ctrl.bootstrap = repo
	ctrl.client = client
	ctrl.resolution = resolution
}
func (ctrl *AccountController) BootstrapLogin(c *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	if c.ShouldBindJSON(&body) != nil || ctrl.bootstrap == nil {
		respondUnauthorized(c)
		return
	}
	ok, err := ctrl.bootstrap.Verify(c.Request.Context(), body.Token)
	if err != nil {
		respondError(c, err)
		return
	}
	if !ok {
		respondUnauthorized(c)
		return
	}
	if err := middleware.SetBootstrapSession(c, true, true); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
func (ctrl *AccountController) accountChoices(c *gin.Context) ([]*application.AccountDTO, error) {
	token := middleware.OIDCAccessToken(c)
	if token == "" || ctrl.client == nil {
		return nil, auth.ErrResHubSession
	}
	normal, err := ctrl.resolution.Resolve(c.Request.Context(), "", token)
	if err != nil {
		return nil, err
	}
	if normal.User == nil {
		return nil, auth.ErrResHubSession
	}
	choices := []*application.AccountDTO{application.NewAccountDTO(normal.User)}
	snapshot, err := ctrl.client.UserInfo(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}
	match, err := ctrl.bootstrap.Matches(c.Request.Context(), snapshot.Issuer, snapshot.Subject)
	if err != nil {
		return nil, err
	}
	if match {
		root, err := ctrl.bootstrap.Current(c.Request.Context())
		if err != nil {
			return nil, err
		}
		choices = append(choices, application.NewAccountDTO(root))
	}
	return choices, nil
}
func (ctrl *AccountController) Accounts(c *gin.Context) {
	choices, err := ctrl.accountChoices(c)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, choices)
}
func (ctrl *AccountController) SelectAccount(c *gin.Context) {
	var body struct {
		ID int `json:"id"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	choices, err := ctrl.accountChoices(c)
	if err != nil {
		respondError(c, err)
		return
	}
	for _, choice := range choices {
		if choice.ID == body.ID {
			if err := middleware.SetBootstrapSession(c, choice.IsBootstrap, false); err != nil {
				respondError(c, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
			return
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
}
