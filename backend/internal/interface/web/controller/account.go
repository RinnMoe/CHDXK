package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
	"net/http"
	"net/url"
)

type AccountController struct {
	query      *application.AccountQueryService
	oidc       *application.ResHubOIDCService
	accounts   *application.LocalAccountService
	resolution *application.AuthResolutionService
}

func NewAccountController(query *application.AccountQueryService, oidc *application.ResHubOIDCService, accounts *application.LocalAccountService, resolution *application.AuthResolutionService) *AccountController {
	return &AccountController{query, oidc, accounts, resolution}
}
func (ctrl *AccountController) OIDCStart(c *gin.Context) {
	returnPath := application.SafeReturnPath(c.Query("return_to"))
	id, version := 0, 0
	if c.Query("purpose") == "bind" {
		id, version = middleware.AccountSessionID(c), middleware.AccountSessionVersion(c)
		resolved, err := ctrl.resolution.ResolveSession(c.Request.Context(), "", middleware.OIDCAccessToken(c), id, version)
		if err != nil || resolved.User == nil {
			respondUnauthorized(c)
			return
		}
		if resolved.User.MustChangePassword {
			respondError(c, auth.ErrPasswordRequired)
			return
		}
	}
	if err := middleware.SetOIDCBindingTarget(c, id, version); err != nil {
		respondError(c, err)
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
		ctrl.redirectOIDCError(c, returnPath, "oidc_unavailable")
		return
	}
	c.Redirect(http.StatusFound, location)
}
func (ctrl *AccountController) OIDCCallback(c *gin.Context) {
	state, verifier, nonce, returnPath, ok := middleware.TakeOIDCState(c, c.Query("state"))
	if !ok || state == "" || c.Query("error") != "" {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	bindingID, bindingVersion := middleware.TakeOIDCBindingTarget(c)
	identity, session, err := ctrl.oidc.CompleteIdentity(c.Request.Context(), c.Query("code"), verifier, nonce)
	if err != nil || identity == nil || session.AccessToken == "" || session.IDToken == "" {
		ctrl.redirectOIDCError(c, returnPath, "oidc_failed")
		return
	}
	var choices []*auth.User
	if bindingID > 0 {
		actor, err := ctrl.accounts.Repo.FindByID(c.Request.Context(), bindingID)
		if err != nil || actor == nil || actor.AuthVersion != bindingVersion || actor.IsSuspended() || actor.MustChangePassword {
			respondUnauthorized(c)
			return
		}
		if err := ctrl.accounts.Repo.BindAccount(c.Request.Context(), bindingID, identity); err != nil {
			respondError(c, err)
			return
		}
		actor, err = ctrl.accounts.Repo.FindByID(c.Request.Context(), bindingID)
		if err != nil {
			respondError(c, err)
			return
		}
		choices = []*auth.User{actor}
		returnPath = "/settings"
		_ = ctrl.accounts.Record(c.Request.Context(), actor, "user.identity.bind", actor.ID, nil)
	} else {
		choices, err = ctrl.accounts.Repo.ResHubAccounts(c.Request.Context(), identity, true)
		if err != nil {
			respondError(c, err)
			return
		}
	}
	if err := middleware.SetSessionOIDC(c, session.AccessToken, session.IDToken); err != nil {
		respondError(c, err)
		return
	}
	if len(choices) > 1 {
		q := url.Values{"redirect": []string{application.SafeReturnPath(returnPath)}}
		c.Redirect(302, "/login/accounts?"+q.Encode())
		return
	}
	if len(choices) != 1 {
		respondUnauthorized(c)
		return
	}
	if err := auth.EnsureUserActive(c.Request.Context(), choices[0]); err != nil {
		_ = middleware.ClearSession(c)
		respondError(c, err)
		return
	}
	if err := middleware.SetAccountSession(c, choices[0].ID, choices[0].AuthVersion, false); err != nil {
		respondError(c, err)
		return
	}
	if choices[0].MustChangePassword {
		returnPath = "/change-password"
	}
	c.Redirect(302, application.SafeReturnPath(returnPath))
}
func (ctrl *AccountController) redirectOIDCError(c *gin.Context, path, code string) {
	query := url.Values{"error": []string{code}, "redirect": []string{application.SafeReturnPath(path)}}
	c.Redirect(302, "/login?"+query.Encode())
}
func (ctrl *AccountController) Logout(c *gin.Context) {
	if err := middleware.ClearSession(c); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
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
	c.JSON(200, dto)
}
func (ctrl *AccountController) accountChoices(c *gin.Context) ([]*auth.User, error) {
	if middleware.NativePendingExpired(c) {
		_ = middleware.ClearSession(c)
		return nil, auth.ErrResHubSession
	}
	token := middleware.OIDCAccessToken(c)
	if token == "" {
		return nil, auth.ErrResHubSession
	}
	identity, err := ctrl.oidc.Identity(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}
	return ctrl.accounts.Repo.ResHubAccounts(c.Request.Context(), identity, false)
}
func (ctrl *AccountController) Accounts(c *gin.Context) {
	choices, err := ctrl.accountChoices(c)
	if err != nil {
		respondError(c, err)
		return
	}
	dtos := []*application.AccountDTO{}
	for _, u := range choices {
		dtos = append(dtos, application.NewAccountDTO(u))
	}
	c.JSON(200, dtos)
}
func (ctrl *AccountController) SelectAccount(c *gin.Context) {
	var body struct {
		ID int `json:"id"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请选择一个账号"})
		return
	}
	choices, err := ctrl.accountChoices(c)
	if err != nil {
		respondError(c, err)
		return
	}
	for _, u := range choices {
		if u.ID == body.ID {
			if err := auth.EnsureUserActive(c.Request.Context(), u); err != nil {
				respondError(c, err)
				return
			}
			if err := middleware.SetAccountSession(c, u.ID, u.AuthVersion, true); err != nil {
				respondError(c, err)
				return
			}
			c.JSON(200, application.NewAccountDTO(u))
			return
		}
	}
	c.JSON(403, gin.H{"error": "不能选择该账号"})
}
