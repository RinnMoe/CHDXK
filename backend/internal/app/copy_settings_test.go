//go:build test

package app

import (
	"context"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
	"jcourse/pkg/sitecopy"
	"testing"
)

func TestCopySettingsSavePreservesWhitespaceAndRejectsMissingParameters(t *testing.T) {
	service := newSystemSettingsService(setting.NewMockSystemRepository(), course.NewMockCourseRepository())
	ctx := context.Background()
	const brandKey = "copy.user.config.brand.743e9916a4b2"
	saved, err := service.Save(ctx, brandKey, "  频道课程榜  ")
	if err != nil || saved.Value != "  频道课程榜  " {
		t.Fatalf("whitespace lost: %+v, %v", saved, err)
	}
	for key, entry := range sitecopy.Entries() {
		if len(sitecopy.Parameters(entry.Text)) > 0 {
			if _, err := service.Save(ctx, key, "缺少内容标记"); err == nil {
				t.Fatal("invalid template accepted")
			}
			break
		}
	}
	if _, ok := service.Definition(setting.SystemSettingKeyReviewFrequencySuspendDuration); ok {
		t.Fatal("local account suspension setting exposed")
	}
}
