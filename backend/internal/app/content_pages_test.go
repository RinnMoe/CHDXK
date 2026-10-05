//go:build test

package app

import (
	"context"
	"jcourse/internal/application"
	"jcourse/internal/domain/course"
	"jcourse/internal/domain/setting"
	"jcourse/pkg/sitepages"
	"testing"
)

func TestDraftPagesAreNotPublic(t *testing.T) {
	service := newSystemSettingsService(setting.NewMockSystemRepository(), course.NewMockCourseRepository())
	ctx := context.Background()
	_, err := service.Save(ctx, sitepages.Key, `[{"slug":"links","title":"友链","content":"公开","published":true,"footer":true},{"slug":"draft","title":"草稿","content":"仅后台可见","published":false,"footer":true}]`)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := application.NewSystemSettingsQueryService(service).Pages(ctx)
	if err != nil || len(pages) != 1 || pages[0].Slug != "links" {
		t.Fatalf("public pages: %+v %v", pages, err)
	}
}
