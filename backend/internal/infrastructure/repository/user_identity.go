package repository

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/point"
	"strings"
	"time"
)

type UserIdentityRepository struct {
	db     *gorm.DB
	issuer string
}

func NewUserIdentityRepository(db *gorm.DB, issuer string) *UserIdentityRepository {
	return &UserIdentityRepository{db: db, issuer: strings.TrimRight(strings.TrimSpace(issuer), "/")}
}
func loadAccount(db *gorm.DB, id int) (*auth.User, error) {
	var e UserEntity
	if err := db.Where("id=? AND deleted_at IS NULL", id).Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	u := newUserDomain(&e)
	var i UserIdentityEntity
	result := db.Where("user_id=?", id).Find(&i)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		u.Subject = i.Subject
		u.IdentityRole = i.IdentityRole
		u.IdentityType = i.IdentityType
		u.BoundNickname = i.Username
		u.HasPointsAccount = i.HasPointsAccount
		u.Email = i.Email
		u.Role = auth.EffectiveRole(e.Role, i.IdentityRole)
	}
	return &u, nil
}
func (r *UserIdentityRepository) FindByID(ctx context.Context, id int) (*auth.User, error) {
	return loadAccount(r.db.WithContext(ctx), id)
}
func (r *UserIdentityRepository) validateIdentity(s *auth.ResHubIdentity) error {
	if s == nil || strings.TrimSpace(s.Subject) == "" || r.issuer == "" || strings.TrimRight(s.Issuer, "/") != r.issuer {
		return auth.ErrIdentityNotFound
	}
	return nil
}
func (r *UserIdentityRepository) BindAccount(ctx context.Context, id int, s *auth.ResHubIdentity) error {
	if err := r.validateIdentity(s); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var u UserEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&u).Error; err != nil {
			return err
		}
		if err := saveIdentity(tx, id, s); err != nil {
			return err
		}
		return tx.Model(&u).Update("auth_version", gorm.Expr("auth_version+1")).Error
	})
}
func saveIdentity(tx *gorm.DB, id int, s *auth.ResHubIdentity) error {
	now := time.Now()
	i := UserIdentityEntity{UserID: id, Issuer: strings.TrimRight(s.Issuer, "/"), Subject: s.Subject, Username: s.Username, Email: s.Email, IdentityRole: s.IdentityRole, IdentityType: s.IdentityType, HasPointsAccount: s.HasPointsAccount, CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"issuer", "subject", "username", "email", "identity_role", "identity_type", "has_points_account", "updated_at"})}).Create(&i).Error
}
func (r *UserIdentityRepository) ResHubAccounts(ctx context.Context, s *auth.ResHubIdentity, create bool) ([]*auth.User, error) {
	if err := r.validateIdentity(s); err != nil {
		return nil, err
	}
	users := []*auth.User{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize first-login creation for each external subject, including across API replicas.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 21))", r.issuer+":"+s.Subject).Error; err != nil {
			return err
		}
		var ids []int
		if err := tx.Table("user_identities i").Joins("JOIN users u ON u.id=i.user_id").Where("i.issuer=? AND i.subject=? AND u.deleted_at IS NULL", r.issuer, s.Subject).Order("u.id").Pluck("u.id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 && create {
			suffix, err := randomAccountSecret()
			if err != nil {
				return err
			}
			now := time.Now()
			e := UserEntity{Username: "hub_" + strings.ToLower(suffix[:12]), Nickname: s.Username, Role: auth.RoleUser, PasswordHash: "", Permissions: "[]", AuthVersion: 1, CreatedAt: now, LastSeenAt: now}
			if err := tx.Create(&e).Error; err != nil {
				return err
			}
			ids = append(ids, e.ID)
		}
		for _, id := range ids {
			if err := saveIdentity(tx, id, s); err != nil {
				return err
			}
			u, err := loadAccount(tx, id)
			if err != nil {
				return err
			}
			users = append(users, u)
		}
		return nil
	})
	return users, err
}
func (r *UserIdentityRepository) Upsert(ctx context.Context, s *auth.ResHubIdentity) (*auth.User, error) {
	users, err := r.ResHubAccounts(ctx, s, true)
	if err != nil {
		return nil, err
	}
	return users[0], nil
}
func (r *UserIdentityRepository) UnbindAccount(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e UserEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).Take(&e).Error; err != nil {
			return err
		}
		if e.PasswordHash == "" {
			return auth.ErrLocalLogin
		}
		if err := tx.Where("user_id=?", id).Delete(&UserIdentityEntity{}).Error; err != nil {
			return err
		}
		return tx.Model(&e).Update("auth_version", gorm.Expr("auth_version+1")).Error
	})
}
func (r *UserIdentityRepository) SubjectByUserID(ctx context.Context, id int) (string, error) {
	u, err := r.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if u == nil || u.Subject == "" || !u.HasPointsAccount {
		return "", point.ErrNoPointAccount
	}
	var i UserIdentityEntity
	if err := r.db.WithContext(ctx).Where("user_id=? AND issuer=?", id, r.issuer).Take(&i).Error; err != nil {
		return "", err
	}
	return i.Subject, nil
}
func permissionsJSON(values []string) string {
	if values == nil {
		values = []string{}
	}
	b, _ := json.Marshal(values)
	return string(b)
}

var _ auth.IdentityProjectionRepository = (*UserIdentityRepository)(nil)
var _ auth.AccountRepository = (*UserIdentityRepository)(nil)
