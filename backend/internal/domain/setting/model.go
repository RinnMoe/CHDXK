package setting

import (
	"context"

	"jcourse/pkg/apperr"
)

var (
	ErrInvalidCurrentSemester    = apperr.ErrInvalidCurrentSemester
	ErrInvalidSystemSettingKey   = apperr.BadRequest("系统设置项无效")
	ErrInvalidSystemSettingValue = apperr.BadRequest("系统设置值无效")
)

type SystemSetting struct {
	Key   string
	Value string
}

type EffectiveSystemSetting struct {
	Definition Definition
	Value      string
}

type SystemRepository interface {
	List(ctx context.Context) ([]SystemSetting, error)
	Get(ctx context.Context, key string) (*SystemSetting, error)
	Save(ctx context.Context, setting *SystemSetting) error
}
