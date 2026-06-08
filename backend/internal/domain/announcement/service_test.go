package announcement

import (
	"errors"
	"testing"
	"time"
)

func TestAnnouncementUpdateValidation(t *testing.T) {
	now := time.Now()
	base := SaveInput{
		Title:     "公告",
		Body:      "内容",
		ShowStart: now,
		ShowEnd:   now.Add(time.Hour),
	}

	tests := []struct {
		name string
		edit func(*SaveInput)
		want error
	}{
		{name: "content required", edit: func(input *SaveInput) { input.Title, input.Body = " ", " " }, want: ErrAnnouncementContentEmpty},
		{name: "invalid range", edit: func(input *SaveInput) { input.ShowEnd = input.ShowStart }, want: ErrAnnouncementInvalidRange},
		{name: "invalid link", edit: func(input *SaveInput) { input.LinkURL = "javascript:alert(1)" }, want: ErrAnnouncementInvalidLink},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := base
			tt.edit(&input)
			item := &Announcement{}
			if err := item.Update(input); !errors.Is(err, tt.want) {
				t.Fatalf("Update error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestAnnouncementUpdateAllowsTitleOnlyOrBodyOnly(t *testing.T) {
	now := time.Now()
	for _, input := range []SaveInput{
		{Title: "公告", Body: "", ShowStart: now, ShowEnd: now.Add(time.Hour)},
		{Title: "", Body: "内容", ShowStart: now, ShowEnd: now.Add(time.Hour)},
	} {
		item := &Announcement{}
		if err := item.Update(input); err != nil {
			t.Fatalf("Update(%+v) error = %v", input, err)
		}
	}
}

func TestAnnouncementUpdateTrimsAndAcceptsHTTPSLink(t *testing.T) {
	now := time.Now()
	item := &Announcement{}
	err := item.Update(SaveInput{
		Title:     " 公告 ",
		Body:      " 内容 ",
		ShowStart: now,
		ShowEnd:   now.Add(time.Hour),
		LinkURL:   " https://example.com/path?q=1 ",
		LinkTitle: " 查看详情 ",
	})
	if err != nil {
		t.Fatalf("Update error = %v", err)
	}
	if item.Title != "公告" || item.Body != "内容" || item.LinkTitle != "查看详情" {
		t.Fatalf("trimmed fields mismatch: %+v", item)
	}
	if item.LinkURL != "https://example.com/path?q=1" {
		t.Fatalf("LinkURL = %q", item.LinkURL)
	}
}
