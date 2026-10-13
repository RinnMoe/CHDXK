package application

import (
	"context"
	"strconv"

	"jcourse/internal/domain/setting"
	"jcourse/pkg/apperr"
)

const (
	ImportConcurrency   = "import.concurrency"
	ImportSchoolEnabled = "import.school.enabled"
	ImportManualEnabled = "import.manual.enabled"
	ImportAdminEnabled  = "import.admin.enabled"
	ImportMyCHUVisible  = "import.mychu.visible"
)

func ImportSettingDefinitions() []setting.Definition {
	defs := []setting.Definition{{Key: ImportConcurrency, Group: "import", Label: "个人导入并发数", Description: "1～64。保存后立即生效；降低时等待已有任务结束，增加时自动接纳排队任务。", Type: setting.ValueTypeInt, DefaultValue: "8", Validator: setting.ValueValidatorFunc(func(_ context.Context, value string) error {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 64 {
			return setting.ErrInvalidSystemSettingValue
		}
		return nil
	})}}
	for _, item := range []struct{ key, label, description string }{
		{ImportSchoolEnabled, "教务扫码导入", "允许用户扫码导入个人课程。关闭后不再接受新任务，已提交任务继续完成。"},
		{ImportManualEnabled, "手动添加课程", "允许用户选择已有课程和教学班添加选课记录。关闭后仍可编辑或删除已有记录。"},
		{ImportAdminEnabled, "管理员教务同步", "允许管理员从教务批量同步开课记录。关闭后不再接受新任务，已提交任务继续完成。"},
		{ImportMyCHUVisible, "显示 MyCHU 预留入口", "仅控制预留入口显示；MyCHU 导入功能尚未开放。"},
	} {
		defs = append(defs, setting.Definition{Key: item.key, Group: "import", Label: item.label, Description: item.description, Type: setting.ValueTypeBool, DefaultValue: "true", Public: true})
	}
	return defs
}

func requireImportEnabled(ctx context.Context, settings *setting.SystemSettingsService, key string) error {
	if settings == nil {
		return nil
	}
	snapshot, err := settings.Snapshot(ctx)
	if err != nil {
		return err
	}
	if !snapshot.Bool(key) {
		return apperr.Forbidden("该导入方式已关闭")
	}
	return nil
}
