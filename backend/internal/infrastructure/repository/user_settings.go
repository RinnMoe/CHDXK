package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/setting"
)

type UserSettingsRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUserSettingsRepository(db *gorm.DB, cache ...*redis.Client) *UserSettingsRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &UserSettingsRepository{db: db, cache: client}
}

func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID int) (*setting.UserSettings, error) {
	key := cacheKey("user_setting", userID)
	if cached, ok := cacheGetJSON[setting.UserSettings](ctx, r.cache, key); ok {
		return cached, nil
	}

	var e UserSettingsEntity
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	settings := &setting.UserSettings{UserID: e.UserID, CurrentSemester: e.CurrentSemester}
	cacheSetJSON(ctx, r.cache, key, settings)
	return settings, nil
}

func (r *UserSettingsRepository) Save(ctx context.Context, settings *setting.UserSettings) error {
	now := time.Now()
	e := UserSettingsEntity{
		UserID:          settings.UserID,
		CurrentSemester: settings.CurrentSemester,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"current_semester": e.CurrentSemester,
			"updated_at":       e.UpdatedAt,
		}),
	}).Create(&e).Error
	if err != nil {
		return err
	}
	cacheDelete(ctx, r.cache, cacheKey("user_setting", settings.UserID))
	return nil
}

var _ setting.Repository = (*UserSettingsRepository)(nil)
