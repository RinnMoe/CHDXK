package setting

import (
	"context"

	"jcourse/pkg/apperr"
)

var (
	ErrInvalidCurrentSemester  = apperr.ErrInvalidCurrentSemester
	ErrInvalidSystemSettingKey = apperr.BadRequest("系统设置项无效")
)

const SystemSettingKeyCurrentSemester = "current_semester"

var registeredSystemSettingKeys = map[string]struct{}{
	SystemSettingKeyCurrentSemester: {},
}

func IsRegisteredSystemSettingKey(key string) bool {
	_, ok := registeredSystemSettingKeys[key]
	return ok
}

type UserSettings struct {
	UserID          int
	CurrentSemester string
}

type SystemSetting struct {
	Key   string
	Value string
}

type Repository interface {
	GetByUserID(ctx context.Context, userID int) (*UserSettings, error)
	Save(ctx context.Context, settings *UserSettings) error
}

type SystemRepository interface {
	List(ctx context.Context) ([]SystemSetting, error)
	Get(ctx context.Context, key string) (*SystemSetting, error)
	Save(ctx context.Context, setting *SystemSetting) error
}
