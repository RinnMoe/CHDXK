//go:build test

package app

import (
	"context"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
	"jcourse/pkg/sitecopy"
	"testing"
)

func TestCopySettingsPreserveWhitespaceAndExcludeGenericCopy(t *testing.T) {
	service := newSystemSettingsService(setting.NewMockSystemRepository(), course.NewMockCourseRepository())
	ctx := context.Background()
	const brandKey = "copy.user.config.brand.743e9916a4b2"
	saved, err := service.Save(ctx, brandKey, "  频道课程榜  ")
	if err != nil || saved.Value != "  频道课程榜  " {
		t.Fatalf("whitespace lost: %+v, %v", saved, err)
	}
	for key, entry := range sitecopy.Entries() {
		_, registered := service.Definition(key)
		if registered != entry.Editable {
			t.Fatalf("editable %s = %v, registered = %v", key, entry.Editable, registered)
		}
		if !entry.Editable {
			if _, err := service.Save(ctx, key, "任意修改"); err == nil {
				t.Fatalf("fixed copy accepted: %s", key)
			}
		}
	}
	if _, ok := service.Definition(setting.SystemSettingKeyReviewFrequencySuspendDuration); ok {
		t.Fatal("local account suspension setting exposed")
	}
}
