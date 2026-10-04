package app

import (
	"context"
	"fmt"
	"jcourse/internal/domain/setting"
	"jcourse/pkg/sitecopy"
	"strings"
)

func copyDefinitions() []setting.Definition {
	entries := sitecopy.Entries()
	defs := make([]setting.Definition, 0, len(entries))
	for key, entry := range entries {
		audience := "普通用户"
		if entry.Audience == "admin" {
			audience = "普通管理员"
		}
		if entry.Audience == "technical" {
			audience = "技术管理员"
		}
		parameters := sitecopy.Parameters(entry.Text)
		description := fmt.Sprintf("面向%s的文案。保存后刷新页面生效。", audience)
		if len(parameters) > 0 {
			description += " 请保留这些内容标记：" + strings.Join(parameters, "、") + "；它们会自动显示实际内容。"
		}
		label := []rune(strings.TrimSpace(entry.Text))
		if key == "copy.brand.feedback_email" {
			label = []rune("联系邮箱")
		}
		if len(label) > 60 {
			label = append(label[:60], []rune("…")...)
		}
		defs = append(defs, setting.Definition{
			Key: key, Group: "copy_" + entry.Audience, Label: string(label), Description: description,
			Type: setting.ValueTypeString, DefaultValue: entry.Text, Public: true, PreserveWhitespace: true,
			Validator: setting.ValueValidatorFunc(func(_ context.Context, value string) error {
				if !sitecopy.ValidValue(key, value) {
					return setting.ErrInvalidSystemSettingValue
				}
				return nil
			}),
		})
	}
	return defs
}
