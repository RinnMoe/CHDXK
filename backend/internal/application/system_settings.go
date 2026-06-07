package application

import "jcourse/internal/domain/setting"

type SystemSettingDTO struct {
	Key             string `json:"key"`
	Value           string `json:"value"`
	DefaultValue    string `json:"default_value"`
	Type            string `json:"type"`
	Group           string `json:"group"`
	Label           string `json:"label"`
	Description     string `json:"description"`
	Public          bool   `json:"public"`
	Secret          bool   `json:"secret"`
	RequiresRestart bool   `json:"requires_restart"`
}

type UpdateSystemSettingCommand struct {
	Value string `json:"value"`
}

func newSystemSettingDTO(item setting.EffectiveSystemSetting) SystemSettingDTO {
	def := item.Definition
	return SystemSettingDTO{
		Key:             def.Key,
		Value:           item.Value,
		DefaultValue:    def.DefaultValue,
		Type:            string(def.Type),
		Group:           def.Group,
		Label:           def.Label,
		Description:     def.Description,
		Public:          def.Public,
		Secret:          def.Secret,
		RequiresRestart: def.RequiresRestart,
	}
}
