package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"jcourse/internal/domain/auth"
	"jcourse/pkg/apperr"
	"strings"
	"time"
)

func randomAccountSecret() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), err
}
func (r *UserIdentityRepository) Initialize(ctx context.Context) (*auth.InitialCredential, error) {
	var result *auth.InitialCredential
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(71818001)").Error; err != nil {
			return err
		}
		var e UserEntity
		err := tx.Where("id=1").Take(&e).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && e.PasswordHash != "" {
			return nil
		}
		password, err := randomAccountSecret()
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if e.ID == 0 {
			now := time.Now()
			e = UserEntity{ID: 1, Username: "admin", Nickname: "初始超级管理员", Role: auth.RoleSuperAdmin, PasswordHash: string(hash), MustChangePassword: true, AuthVersion: 1, Permissions: "[]", CreatedAt: now, LastSeenAt: now}
			if err := tx.Create(&e).Error; err != nil {
				return err
			}
			if err := tx.Exec("SELECT setval(pg_get_serial_sequence('users','id'), GREATEST((SELECT MAX(id) FROM users), nextval(pg_get_serial_sequence('users','id'))), true)").Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&e).Updates(map[string]any{"password_hash": string(hash), "must_change_password": true, "role": auth.RoleSuperAdmin, "auth_version": gorm.Expr("auth_version+1")}).Error; err != nil {
				return err
			}
		}
		result = &auth.InitialCredential{Username: e.Username, Password: password}
		return nil
	})
	return result, err
}
func accountWriteError(err error) error {
	if err != nil && (strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key")) {
		return apperr.BadRequest("该账号名已被使用")
	}
	return err
}
func (r *UserIdentityRepository) FindCredential(ctx context.Context, username string) (*auth.AccountCredential, error) {
	var e UserEntity
	if err := r.db.WithContext(ctx).Where("username=? AND deleted_at IS NULL", username).Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	u, err := r.FindByID(ctx, e.ID)
	return &auth.AccountCredential{User: u, PasswordHash: e.PasswordHash}, err
}
func (r *UserIdentityRepository) CreateLocal(ctx context.Context, username, nickname, hash string, force bool) (*auth.User, error) {
	now := time.Now()
	e := UserEntity{Username: username, Nickname: nickname, PasswordHash: hash, MustChangePassword: force, Role: auth.RoleUser, Permissions: "[]", AuthVersion: 1, CreatedAt: now, LastSeenAt: now}
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return nil, accountWriteError(err)
	}
	return r.FindByID(ctx, e.ID)
}
func (r *UserIdentityRepository) SetPassword(ctx context.Context, id int, hash string, force bool) error {
	return r.db.WithContext(ctx).Model(&UserEntity{}).Where("id=? AND deleted_at IS NULL", id).Updates(map[string]any{"password_hash": hash, "must_change_password": force, "auth_version": gorm.Expr("auth_version+1")}).Error
}
func (r *UserIdentityRepository) UpdateProfile(ctx context.Context, id int, username, nickname string) error {
	return accountWriteError(r.db.WithContext(ctx).Model(&UserEntity{}).Where("id=? AND deleted_at IS NULL", id).Updates(map[string]any{"username": username, "nickname": nickname}).Error)
}
func (r *UserIdentityRepository) ListAccounts(ctx context.Context, f auth.AccountFilter) ([]*auth.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&UserEntity{}).Where("deleted_at IS NULL")
	if f.Query != "" {
		value := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(f.Query) + "%"
		q = q.Where("username ILIKE ? OR nickname ILIKE ?", value, value)
	}
	if f.Role != "" {
		if f.Role == auth.RoleAdmin {
			q = q.Where("role='admin' OR (role<>'super_admin' AND EXISTS (SELECT 1 FROM user_identities i WHERE i.user_id=users.id AND i.identity_role IN ('admin','administrator','superadmin','super_admin','super-admin')))")
		} else if f.Role == auth.RoleUser {
			q = q.Where("role='user' AND NOT EXISTS (SELECT 1 FROM user_identities i WHERE i.user_id=users.id AND i.identity_role IN ('admin','administrator','superadmin','super_admin','super-admin'))")
		} else {
			q = q.Where("role=?", f.Role)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := max(1, f.Page)
	var ids []int
	if err := q.Order("id").Limit(20).Offset((page-1)*20).Pluck("id", &ids).Error; err != nil {
		return nil, 0, err
	}
	users := []*auth.User{}
	for _, id := range ids {
		u, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		if u != nil {
			users = append(users, u)
		}
	}
	return users, total, nil
}
func (r *UserIdentityRepository) SaveAccount(ctx context.Context, u *auth.User) error {
	return r.db.WithContext(ctx).Model(&UserEntity{}).Where("id=? AND deleted_at IS NULL", u.ID).Updates(map[string]any{"nickname": u.Nickname, "role": u.LocalRole, "permissions": permissionsJSON(u.Permissions), "suspended_at": u.SuspendedAt, "suspend_till": u.SuspendTill, "muted_until": u.MutedUntil, "moderation_reason": u.ModerationReason}).Error
}
func (r *UserIdentityRepository) DeleteAccount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Model(&UserEntity{}).Where("id=?", id).Updates(map[string]any{"deleted_at": time.Now(), "password_hash": "", "auth_version": gorm.Expr("auth_version+1")}).Error
}
