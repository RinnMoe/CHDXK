package application

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"jcourse/internal/domain/audit"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"jcourse/internal/domain/setting"
	"jcourse/pkg/apperr"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const RegistrationSetting = "auth.registration_enabled"

var accountNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,31}$`)
var operationIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-]{8,64}$`)

type LocalAccountService struct {
	Repo     auth.AccountRepository
	settings *setting.SystemSettingsService
	points   point.ExternalPointProvider
	audit    audit.CommandRepository
}

func NewLocalAccountService(repo auth.AccountRepository, settings *setting.SystemSettingsService, points point.ExternalPointProvider, auditRepo audit.CommandRepository) *LocalAccountService {
	return &LocalAccountService{repo, settings, points, auditRepo}
}
func (s *LocalAccountService) RegistrationEnabled(ctx context.Context) (bool, error) {
	snapshot, err := s.settings.Snapshot(ctx)
	if err != nil {
		return false, err
	}
	return snapshot.Bool(RegistrationSetting), nil
}
func validateAccountName(value string) error {
	if !accountNamePattern.MatchString(value) {
		return apperr.BadRequest("账号名须为 3–32 位小写字母、数字、点、横线或下划线")
	}
	return nil
}
func validateNickname(value string) error {
	if utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 64 {
		return apperr.BadRequest("昵称须为 1–64 个字")
	}
	return nil
}
func passwordHash(password string) (string, error) {
	if utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		return "", apperr.BadRequest("密码至少 8 个字符，最多 72 字节")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func (s *LocalAccountService) Login(ctx context.Context, username, password string) (*auth.User, error) {
	c, err := s.Repo.FindCredential(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	if c == nil || c.User == nil || c.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(c.PasswordHash), []byte(password)) != nil {
		return nil, auth.ErrLocalLogin
	}
	if err := auth.EnsureUserActive(ctx, c.User); err != nil {
		return nil, err
	}
	return c.User, nil
}
func (s *LocalAccountService) Create(ctx context.Context, username, nickname, password string, force bool) (*auth.User, error) {
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if err := validateAccountName(username); err != nil {
		return nil, err
	}
	if err := validateNickname(nickname); err != nil {
		return nil, err
	}
	hash, err := passwordHash(password)
	if err != nil {
		return nil, err
	}
	return s.Repo.CreateLocal(ctx, username, nickname, hash, force)
}
func (s *LocalAccountService) Register(ctx context.Context, username, nickname, password string) (*auth.User, error) {
	enabled, err := s.RegistrationEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, apperr.Forbidden("本站暂未开放账号注册")
	}
	return s.Create(ctx, username, nickname, password, false)
}
func (s *LocalAccountService) ChangePassword(ctx context.Context, u *auth.User, old, password string) error {
	if u.HasPassword {
		current, err := s.Login(ctx, u.Username, old)
		if err != nil || current.ID != u.ID {
			return auth.ErrLocalLogin
		}
		if old == password {
			return apperr.BadRequest("新密码不能与原密码相同")
		}
	} else if u.Subject == "" {
		return auth.ErrLocalLogin
	}
	hash, err := passwordHash(password)
	if err != nil {
		return err
	}
	return s.Repo.SetPassword(ctx, u.ID, hash, false)
}
func (s *LocalAccountService) Profile(ctx context.Context, u *auth.User, username, nickname string) error {
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if username != u.Username {
		if err := validateAccountName(username); err != nil {
			return err
		}
	}
	if err := validateNickname(nickname); err != nil {
		return err
	}
	return s.Repo.UpdateProfile(ctx, u.ID, username, nickname)
}
func (s *LocalAccountService) Record(ctx context.Context, actor *auth.User, action string, id int, details audit.Details) error {
	if s.audit == nil {
		return nil
	}
	if details == nil {
		details = audit.Details{}
	}
	return s.audit.Create(ctx, &audit.Log{ActorUserID: actor.ID, Action: action, TargetType: "user", TargetID: fmt.Sprint(id), Details: details, OccurredAt: time.Now()})
}
func (s *LocalAccountService) Target(ctx context.Context, actor *auth.User, id int) (*auth.User, error) {
	u, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperr.NotFound("未找到该用户")
	}
	if id == 1 || id == actor.ID || (u.IsAdmin() && !actor.IsSuperAdmin() && !actor.Can(auth.PermissionRoles)) {
		return nil, apperr.Forbidden("不能通过此操作修改当前账号或受保护的管理员账号")
	}
	for _, p := range u.Permissions {
		if !actor.Can(p) {
			return nil, apperr.Forbidden("不能处理权限高于自己的管理员账号")
		}
	}
	return u, nil
}
func (s *LocalAccountService) SetRole(ctx context.Context, actor *auth.User, id int, role string, permissions []string) error {
	if !actor.Can(auth.PermissionRoles) {
		return apperr.Forbidden("没有用户权限设置权限")
	}
	target, err := s.Target(ctx, actor, id)
	if err != nil {
		return err
	}
	if role != auth.RoleAdmin && role != auth.RoleUser {
		return apperr.BadRequest("只能设置管理员或用户身份")
	}
	seen := map[string]bool{}
	for _, p := range permissions {
		valid := false
		for _, known := range auth.AdminPermissions {
			if p == known {
				valid = true
			}
		}
		if !valid || !actor.Can(p) || seen[p] {
			return apperr.Forbidden("不能授予未拥有的管理权限")
		}
		seen[p] = true
	}
	target.LocalRole = role
	target.Permissions = permissions
	if role == auth.RoleUser && auth.EffectiveRole(role, target.IdentityRole) != auth.RoleAdmin {
		target.Permissions = []string{}
	}
	if err := s.Repo.SaveAccount(ctx, target); err != nil {
		return err
	}
	return s.Record(ctx, actor, "user.role.update", id, audit.Details{"role": role, "permissions": target.Permissions})
}
func (s *LocalAccountService) Moderate(ctx context.Context, actor *auth.User, id int, action, reason string, until *time.Time) error {
	if !actor.IsAdmin() {
		return apperr.Forbidden("没有内容管理权限")
	}
	target, err := s.Target(ctx, actor, id)
	if err != nil {
		return err
	}
	if until != nil && !until.After(time.Now()) {
		return apperr.BadRequest("结束时间须晚于当前时间")
	}
	switch action {
	case "ban":
		now := time.Now()
		target.SuspendedAt = &now
		target.SuspendTill = until
	case "unban":
		target.ClearSuspension()
	case "mute":
		if until == nil {
			v := time.Now().AddDate(100, 0, 0)
			until = &v
		}
		target.MutedUntil = until
	case "unmute":
		target.MutedUntil = nil
	default:
		return apperr.BadRequest("无效的处理方式")
	}
	if len(reason) > 500 {
		return apperr.BadRequest("处理原因过长")
	}
	target.ModerationReason = reason
	if err := s.Repo.SaveAccount(ctx, target); err != nil {
		return err
	}
	return s.Record(ctx, actor, "user."+action, id, audit.Details{"reason": reason, "until": until})
}
func (s *LocalAccountService) AdjustPoints(ctx context.Context, actor *auth.User, id, amount int, reason, key string) (point.ExternalOperation, error) {
	if !actor.Can(auth.PermissionPoints) {
		return point.ExternalOperation{}, apperr.Forbidden("没有积分管理权限")
	}
	if amount == 0 || amount < -1000000 || amount > 1000000 || !operationIDPattern.MatchString(key) || strings.TrimSpace(reason) == "" || len(reason) > 200 {
		return point.ExternalOperation{}, apperr.BadRequest("请填写有效的积分数量和原因")
	}
	target, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return point.ExternalOperation{}, err
	}
	if target == nil || target.Subject == "" || !target.HasPointsAccount {
		return point.ExternalOperation{}, point.ErrNoPointAccount
	}
	operation, err := s.points.Change(ctx, target.Subject, amount, "CHDXK："+reason, fmt.Sprintf("chdxk:admin:%d:%s", actor.ID, key))
	if err != nil {
		return operation, err
	}
	err = s.Record(ctx, actor, "user.points.adjust", id, audit.Details{"amount": amount, "reason": reason, "operation_id": operation.ID})
	return operation, err
}

func (s *LocalAccountService) ResetPassword(ctx context.Context, u *auth.User, password string) error {
	hash, err := passwordHash(password)
	if err != nil {
		return err
	}
	return s.Repo.SetPassword(ctx, u.ID, hash, true)
}
