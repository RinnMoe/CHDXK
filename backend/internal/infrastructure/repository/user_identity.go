package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/auth"
)

const resHubIssuer = "res-hub"

type UserIdentityRepository struct {
	db *gorm.DB
}

func NewUserIdentityRepository(db *gorm.DB) *UserIdentityRepository {
	return &UserIdentityRepository{db: db}
}

func (r *UserIdentityRepository) Upsert(ctx context.Context, snapshot *auth.ResHubIdentity) (*auth.User, error) {
	if snapshot == nil || strings.TrimSpace(snapshot.Subject) == "" {
		return nil, auth.ErrIdentityNotFound
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		result, err := r.upsertOnce(ctx, snapshot)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Two first requests for a new subject can race before the unique
		// issuer/subject index is visible. If the competing transaction won,
		// retry through the update path once; other failures are returned as-is.
		var existing UserIdentityEntity
		if findErr := r.db.WithContext(ctx).
			Where("issuer = ? AND subject = ?", resHubIssuer, snapshot.Subject).
			Take(&existing).Error; findErr != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

func (r *UserIdentityRepository) upsertOnce(ctx context.Context, snapshot *auth.ResHubIdentity) (*auth.User, error) {
	var result *auth.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var identity UserIdentityEntity
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("issuer = ? AND subject = ?", resHubIssuer, snapshot.Subject).
			Take(&identity).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user := UserEntity{
				Username:     internalUsername(snapshot.Subject),
				Role:         normalizeRole(snapshot.EffectiveRole),
				PasswordHash: "",
				CreatedAt:    now,
				LastSeenAt:   now,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			identity = UserIdentityEntity{
				UserID:       user.ID,
				Issuer:       resHubIssuer,
				Subject:      snapshot.Subject,
				Username:     snapshot.Username,
				Email:        snapshot.Email,
				IdentityRole: snapshot.IdentityRole,
				IdentityType: snapshot.IdentityType,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := tx.Create(&identity).Error; err != nil {
				return err
			}
		} else {
			identity.Username = snapshot.Username
			identity.Email = snapshot.Email
			identity.IdentityRole = snapshot.IdentityRole
			identity.IdentityType = snapshot.IdentityType
			identity.UpdatedAt = now
			if err := tx.Save(&identity).Error; err != nil {
				return err
			}
			if err := tx.Model(&UserEntity{}).Where("id = ?", identity.UserID).Updates(map[string]any{
				"last_seen_at": now,
				"role":         normalizeRole(snapshot.EffectiveRole),
			}).Error; err != nil {
				return err
			}
		}

		result = userFromSnapshot(identity.UserID, snapshot)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *UserIdentityRepository) SubjectByUserID(ctx context.Context, userID int) (string, error) {
	if userID <= 0 {
		return "", auth.ErrIdentityNotFound
	}
	var identity UserIdentityEntity
	if err := r.db.WithContext(ctx).Where("issuer = ? AND user_id = ?", resHubIssuer, userID).Take(&identity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", auth.ErrIdentityNotFound
		}
		return "", err
	}
	return identity.Subject, nil
}

func userFromSnapshot(userID int, snapshot *auth.ResHubIdentity) *auth.User {
	return &auth.User{
		ID:               userID,
		Role:             normalizeRole(snapshot.EffectiveRole),
		Subject:          snapshot.Subject,
		Username:         snapshot.Username,
		Email:            snapshot.Email,
		IdentityRole:     snapshot.IdentityRole,
		IdentityType:     snapshot.IdentityType,
		PermissionMode:   snapshot.PermissionMode,
		ViolationCount:   snapshot.ViolationCount,
		InitialPassword:  snapshot.InitialPassword,
		HasPointsAccount: snapshot.HasPointsAccount,
		SessionID:        snapshot.SessionID,
		ExpiresAt:        snapshot.ExpiresAt,
		ImpersonatedBy:   snapshot.ImpersonatedBy,
	}
}

func internalUsername(subject string) string {
	sum := sha256.Sum256([]byte(resHubIssuer + ":" + subject))
	return "sso_" + hex.EncodeToString(sum[:])[:40]
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "super_admin", "super-admin", "superadmin":
		return auth.RoleSuperAdmin
	case "admin", "administrator":
		return auth.RoleAdmin
	default:
		return auth.RoleUser
	}
}

var _ auth.IdentityProjectionRepository = (*UserIdentityRepository)(nil)
