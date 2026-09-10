package controller

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
	"jcourse/pkg/logx"
)

type AccountController struct {
	command *application.AccountCommandService
	query   *application.AccountQueryService
	sso     *application.ResHubSSOService
}

func NewAccountController(command *application.AccountCommandService, query *application.AccountQueryService, sso ...*application.ResHubSSOService) *AccountController {
	ctrl := &AccountController{command: command, query: query}
	if len(sso) > 0 {
		ctrl.sso = sso[0]
	}
	return ctrl
}

func (ctrl *AccountController) SSOStart(c *gin.Context) {
	returnPath := application.SafeReturnPath(c.Query("return_to"))
	if ctrl.sso == nil {
		ctrl.redirectSSOError(c, returnPath, "sso_unavailable")
		return
	}
	state, err := application.NewSSOState()
	if err != nil {
		ctrl.redirectSSOError(c, returnPath, "sso_failed")
		return
	}
	verifier, err := application.NewPKCEVerifier()
	if err != nil {
		ctrl.redirectSSOError(c, returnPath, "sso_failed")
		return
	}
	if err := middleware.SetSSOState(c, state, verifier, returnPath); err != nil {
		ctrl.redirectSSOError(c, returnPath, "sso_failed")
		return
	}
	location, err := ctrl.sso.StartURL(state, verifier)
	if err != nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectSSOError(c, returnPath, "sso_unavailable")
		return
	}
	c.Redirect(http.StatusFound, location)
}

func (ctrl *AccountController) SSOCallback(c *gin.Context) {
	state, verifier, returnPath, ok := middleware.TakeSSOState(c, c.Query("state"))
	if !ok || state == "" {
		_ = middleware.ClearSession(c)
		ctrl.redirectSSOError(c, returnPath, "sso_failed")
		return
	}
	if ctrl.sso == nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectSSOError(c, returnPath, "sso_unavailable")
		return
	}
	if remoteError := c.Query("error"); remoteError != "" {
		_ = middleware.ClearSession(c)
		ctrl.redirectSSOError(c, returnPath, "sso_denied")
		return
	}
	user, token, err := ctrl.sso.Complete(c.Request.Context(), c.Query("code"), verifier)
	if err != nil || user == nil || token == "" {
		_ = middleware.ClearSession(c)
		code := "sso_failed"
		if errors.Is(err, auth.ErrResHubUnavailable) {
			code = "sso_unavailable"
		}
		ctrl.redirectSSOError(c, returnPath, code)
		return
	}
	if err := middleware.SetSessionIntegration(c, user.ID, token); err != nil {
		_ = middleware.ClearSession(c)
		ctrl.redirectSSOError(c, returnPath, "sso_failed")
		return
	}
	c.Redirect(http.StatusFound, application.SafeReturnPath(returnPath))
}

func (ctrl *AccountController) redirectSSOError(c *gin.Context, returnPath, code string) {
	query := url.Values{}
	query.Set("error", code)
	query.Set("redirect", application.SafeReturnPath(returnPath))
	c.Redirect(http.StatusFound, "/login?"+query.Encode())
}

func (ctrl *AccountController) SendRegisterCode(c *gin.Context) {
	var cmd application.SendRegisterCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "send_register_code", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendRegisterCode(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "send_register_code", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) Register(c *gin.Context) {
	var cmd application.RegisterCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "register", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Register(c.Request.Context(), cmd)
	if err != nil {
		logAccountError(c, "register", cmd.Email, err)
		respondError(c, err)
		return
	}
	authHash, err := ctrl.command.SessionAuthHash(c.Request.Context(), u.ID)
	if err != nil {
		logAccountError(c, "register_session_auth_hash", cmd.Email, err)
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUser(c, u.ID, authHash); err != nil {
		logAccountError(c, "register_set_session", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (ctrl *AccountController) Login(c *gin.Context) {
	var cmd application.LoginCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "login", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	u, err := ctrl.command.Login(c.Request.Context(), cmd)
	if err != nil {
		logAccountError(c, "login", cmd.Email, err)
		respondError(c, err)
		return
	}
	authHash, err := ctrl.command.SessionAuthHash(c.Request.Context(), u.ID)
	if err != nil {
		logAccountError(c, "login_session_auth_hash", cmd.Email, err)
		respondError(c, err)
		return
	}
	if err := middleware.SetSessionUser(c, u.ID, authHash); err != nil {
		logAccountError(c, "login_set_session", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (ctrl *AccountController) Logout(c *gin.Context) {
	token := middleware.IntegrationToken(c)
	var remoteErr error
	if token != "" && ctrl.sso != nil {
		remoteErr = ctrl.sso.Logout(c.Request.Context(), token)
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

func (ctrl *AccountController) SendResetCode(c *gin.Context) {
	var cmd application.SendResetCodeCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "send_reset_code", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.SendResetCode(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "send_reset_code", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctrl *AccountController) ResetPassword(c *gin.Context) {
	var cmd application.ResetPasswordCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		logAccountError(c, "reset_password", cmd.Email, err)
		respondBindError(c, err)
		return
	}

	if err := ctrl.command.ResetPassword(c.Request.Context(), cmd); err != nil {
		logAccountError(c, "reset_password", cmd.Email, err)
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func logAccountError(c *gin.Context, action, email string, err error) {
	logx.Warn(c.Request.Context(), "account command failed", "action", action, "email", email, "err", err)
}
