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

type SystemSettingsRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewSystemSettingsRepository(db *gorm.DB, cache ...*redis.Client) *SystemSettingsRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &SystemSettingsRepository{db: db, cache: client}
}

func (r *SystemSettingsRepository) List(ctx context.Context) ([]setting.SystemSetting, error) {
	listCacheKey := cacheKey("system_settings")
	if cached, ok := cacheGetJSON[[]setting.SystemSetting](ctx, r.cache, listCacheKey); ok {
		return *cached, nil
	}

	entities, err := gorm.G[SystemSettingsEntity](r.db).Order("key ASC").Find(ctx)
	if err != nil {
		return nil, err
	}
	settings := make([]setting.SystemSetting, 0, len(entities))
	for i := range entities {
		settings = append(settings, newSystemSettingDomain(&entities[i]))
	}
	cacheSetJSON(ctx, r.cache, listCacheKey, settings)
	return settings, nil
}

func (r *SystemSettingsRepository) Get(ctx context.Context, key string) (*setting.SystemSetting, error) {
	itemCacheKey := cacheKey("system_setting", key)
	if cached, ok := cacheGetJSON[setting.SystemSetting](ctx, r.cache, itemCacheKey); ok {
		return cached, nil
	}

	e, err := gorm.G[SystemSettingsEntity](r.db).Where("key = ?", key).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	setting := newSystemSettingDomain(&e)
	cacheSetJSON(ctx, r.cache, itemCacheKey, setting)
	return &setting, nil
}

func (r *SystemSettingsRepository) Save(ctx context.Context, item *setting.SystemSetting) error {
	now := time.Now()
	e := SystemSettingsEntity{
		Key:       item.Key,
		Value:     item.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      e.Value,
			"updated_at": e.UpdatedAt,
		}),
	}).Create(&e).Error
	if err != nil {
		return err
	}
	cacheDelete(ctx, r.cache, cacheKey("system_settings"), cacheKey("system_setting", item.Key))
	return nil
}

func newSystemSettingDomain(e *SystemSettingsEntity) setting.SystemSetting {
	return setting.SystemSetting{Key: e.Key, Value: e.Value}
}

var _ setting.SystemRepository = (*SystemSettingsRepository)(nil)
