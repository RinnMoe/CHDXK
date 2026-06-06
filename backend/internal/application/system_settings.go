package application

import "jcourse/internal/domain/setting"

type SystemSettingDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdateSystemSettingCommand struct {
	Value string `json:"value"`
}

func newSystemSettingDTO(item setting.SystemSetting) SystemSettingDTO {
	return SystemSettingDTO{Key: item.Key, Value: item.Value}
}
