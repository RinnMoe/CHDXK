package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/auth"
)

type BootstrapAccountEntity struct {
	UserID           int `gorm:"primaryKey;autoIncrement:false"`
	TokenHash        string
	Issuer           string
	Subject          string
	Nickname         string
	HasPointsAccount bool
}

func (BootstrapAccountEntity) TableName() string { return "bootstrap_accounts" }

type BootstrapAccountRepository struct{ db *gorm.DB }

func NewBootstrapAccountRepository(db *gorm.DB) *BootstrapAccountRepository {
	return &BootstrapAccountRepository{db}
}
func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
func (r *BootstrapAccountRepository) Ensure(ctx context.Context, rotate bool) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(71818001)").Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&BootstrapAccountEntity{}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			if !rotate {
				return nil
			}
			created = true
			return tx.Model(&BootstrapAccountEntity{}).Where("user_id=1").Update("token_hash", tokenHash(token)).Error
		}
		now := time.Now()
		user := UserEntity{ID: 1, Username: "bootstrap-owner", Role: auth.RoleSuperAdmin, PasswordHash: "", CreatedAt: now, LastSeenAt: now}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		created = true
		return tx.Create(&BootstrapAccountEntity{UserID: 1, TokenHash: tokenHash(token)}).Error
	})
	if err != nil {
		return "", err
	}
	if !created {
		return "", nil
	}
	return token, nil
}
func (r *BootstrapAccountRepository) Verify(ctx context.Context, token string) (bool, error) {
	var e BootstrapAccountEntity
	if err := r.db.WithContext(ctx).Where("user_id=1").Take(&e).Error; err != nil {
		return false, err
	}
	return len(token) == 43 && subtle.ConstantTimeCompare([]byte(e.TokenHash), []byte(tokenHash(token))) == 1, nil
}
func (r *BootstrapAccountRepository) Current(ctx context.Context) (*auth.User, error) {
	var e BootstrapAccountEntity
	if err := r.db.WithContext(ctx).Where("user_id=1").Take(&e).Error; err != nil {
		return nil, err
	}
	return &auth.User{ID: 1, Role: auth.RoleSuperAdmin, Username: "初始超级管理员", IsBootstrap: true, Subject: e.Subject, HasPointsAccount: e.HasPointsAccount, BoundNickname: e.Nickname}, nil
}
func (r *BootstrapAccountRepository) Bind(ctx context.Context, i *auth.ResHubIdentity) error {
	if i == nil || i.Subject == "" || i.Issuer == "" {
		return auth.ErrResHubSession
	}
	return r.db.WithContext(ctx).Model(&BootstrapAccountEntity{}).Where("user_id=1").Updates(map[string]any{"issuer": i.Issuer, "subject": i.Subject, "nickname": i.Username, "has_points_account": i.HasPointsAccount}).Error
}
func (r *BootstrapAccountRepository) Matches(ctx context.Context, issuer, subject string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&BootstrapAccountEntity{}).Where("user_id=1 AND issuer=? AND subject=?", issuer, subject).Count(&count).Error
	return count == 1 && subject != "", err
}
