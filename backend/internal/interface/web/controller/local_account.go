package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/domain/auth"
	"jcourse/internal/interface/web/middleware"
)

type passwordAccountBody struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (ctrl *AccountController) Options(c *gin.Context) {
	enabled, err := ctrl.accounts.RegistrationEnabled(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"registration_enabled": enabled})
}
func (ctrl *AccountController) loginUser(c *gin.Context, u *auth.User) {
	if err := middleware.SetPasswordSession(c, u.ID, u.AuthVersion); err != nil {
		respondError(c, err)
		return
	}
	dto, err := ctrl.query.CurrentUser(c.Request.Context(), u)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, dto)
}
func (ctrl *AccountController) PasswordLogin(c *gin.Context) {
	var body passwordAccountBody
	if c.ShouldBindJSON(&body) != nil {
		respondError(c, auth.ErrLocalLogin)
		return
	}
	u, err := ctrl.accounts.Login(c.Request.Context(), body.Username, body.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	ctrl.loginUser(c, u)
}
func (ctrl *AccountController) Register(c *gin.Context) {
	var body passwordAccountBody
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写账号、昵称和密码"})
		return
	}
	u, err := ctrl.accounts.Register(c.Request.Context(), body.Username, body.Nickname, body.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	ctrl.loginUser(c, u)
}
func (ctrl *AccountController) ChangePassword(c *gin.Context) {
	var body struct {
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写密码"})
		return
	}
	u := auth.GetUserFromCtx(c.Request.Context())
	if err := ctrl.accounts.ChangePassword(c.Request.Context(), u, body.OldPassword, body.Password); err != nil {
		respondError(c, err)
		return
	}
	fresh, err := ctrl.accounts.Repo.FindByID(c.Request.Context(), u.ID)
	if err != nil {
		respondError(c, err)
		return
	}
	if err := middleware.SetAccountSession(c, u.ID, fresh.AuthVersion, true); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *AccountController) Profile(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写账号和昵称"})
		return
	}
	u := auth.GetUserFromCtx(c.Request.Context())
	if err := ctrl.accounts.Profile(c.Request.Context(), u, body.Username, body.Nickname); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *AccountController) Unbind(c *gin.Context) {
	u := auth.GetUserFromCtx(c.Request.Context())
	if err := ctrl.accounts.Repo.UnbindAccount(c.Request.Context(), u.ID); err != nil {
		respondError(c, err)
		return
	}
	_ = middleware.ClearSession(c)
	_ = ctrl.accounts.Record(c.Request.Context(), u, "user.identity.unbind", u.ID, nil)
	c.JSON(200, gin.H{"message": "ok"})
}
