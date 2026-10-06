package controller

import (
	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"strconv"
	"time"
)

type UserManagementController struct {
	s *application.LocalAccountService
}

func NewUserManagementController(s *application.LocalAccountService) *UserManagementController {
	return &UserManagementController{s}
}
func userID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("userID"))
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "无效的用户编号"})
		return 0, false
	}
	return id, true
}
func (ctrl *UserManagementController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	users, total, err := ctrl.s.Repo.ListAccounts(c.Request.Context(), auth.AccountFilter{Query: c.Query("q"), Role: c.Query("role"), Page: page})
	if err != nil {
		respondError(c, err)
		return
	}
	items := []*application.AccountDTO{}
	for _, u := range users {
		items = append(items, application.NewAccountDTO(u))
	}
	c.JSON(200, gin.H{"items": items, "total": total})
}
func (ctrl *UserManagementController) Create(c *gin.Context) {
	var body passwordAccountBody
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写完整的账号资料"})
		return
	}
	u, err := ctrl.s.Create(c.Request.Context(), body.Username, body.Nickname, body.Password, true)
	if err != nil {
		respondError(c, err)
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	_ = ctrl.s.Record(c.Request.Context(), actor, "user.create", u.ID, nil)
	c.JSON(201, application.NewAccountDTO(u))
}
func (ctrl *UserManagementController) Update(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	target, err := ctrl.s.Target(c.Request.Context(), actor, id)
	if err != nil {
		respondError(c, err)
		return
	}
	var body struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写账号资料"})
		return
	}
	if err := ctrl.s.Profile(c.Request.Context(), target, body.Username, body.Nickname); err != nil {
		respondError(c, err)
		return
	}
	_ = ctrl.s.Record(c.Request.Context(), actor, "user.profile.update", id, audit.Details{"username": body.Username, "nickname": body.Nickname})
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *UserManagementController) Delete(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	if _, err := ctrl.s.Target(c.Request.Context(), actor, id); err != nil {
		respondError(c, err)
		return
	}
	if err := ctrl.s.Repo.DeleteAccount(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	_ = ctrl.s.Record(c.Request.Context(), actor, "user.delete", id, nil)
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *UserManagementController) ResetPassword(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	actor := auth.GetUserFromCtx(c.Request.Context())
	target, err := ctrl.s.Target(c.Request.Context(), actor, id)
	if err != nil {
		respondError(c, err)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写新密码"})
		return
	}
	if err := ctrl.s.ResetPassword(c.Request.Context(), target, body.Password); err != nil {
		respondError(c, err)
		return
	}
	_ = ctrl.s.Record(c.Request.Context(), actor, "user.password.reset", id, nil)
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *UserManagementController) Role(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Role        string   `json:"role"`
		Permissions []string `json:"permissions"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写用户权限"})
		return
	}
	if err := ctrl.s.SetRole(c.Request.Context(), auth.GetUserFromCtx(c.Request.Context()), id, body.Role, body.Permissions); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *UserManagementController) Moderate(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Action string     `json:"action"`
		Reason string     `json:"reason"`
		Until  *time.Time `json:"until"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写处理方式"})
		return
	}
	if err := ctrl.s.Moderate(c.Request.Context(), auth.GetUserFromCtx(c.Request.Context()), id, body.Action, body.Reason, body.Until); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
func (ctrl *UserManagementController) Points(c *gin.Context) {
	id, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Amount    int    `json:"amount"`
		Reason    string `json:"reason"`
		RequestID string `json:"request_id"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "请填写积分数量和原因"})
		return
	}
	operation, err := ctrl.s.AdjustPoints(c.Request.Context(), auth.GetUserFromCtx(c.Request.Context()), id, body.Amount, body.Reason, body.RequestID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"balance": operation.BalanceAfter, "operation_id": operation.ID})
}
