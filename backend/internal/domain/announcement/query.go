package announcement

import (
	"context"
	"net/url"
	"strings"
	"time"

	"jcourse/pkg/apperr"
)

var (
	ErrAnnouncementNotFound     = apperr.NotFound("公告不存在")
	ErrAnnouncementContentEmpty = apperr.BadRequest("公告标题和内容不能同时为空")
	ErrAnnouncementInvalidRange = apperr.BadRequest("公告展示时间范围无效")
	ErrAnnouncementInvalidLink  = apperr.BadRequest("公告外链无效")
)

type Announcement struct {
	ID        int
	Title     string
	Body      string
	Priority  int
	ShowStart time.Time
	ShowEnd   time.Time
	LinkURL   string
	LinkTitle string
	CreatedAt time.Time
}

type SaveInput struct {
	Title     string
	Body      string
	Priority  int
	ShowStart time.Time
	ShowEnd   time.Time
	LinkURL   string
	LinkTitle string
}

type AnnouncementView struct {
	ID        int
	Title     string
	Body      string
	Priority  int
	ShowStart time.Time
	ShowEnd   time.Time
	LinkURL   string
	LinkTitle string
	CreatedAt time.Time
}

type AnnouncementQuery interface {
	FindActive(ctx context.Context) ([]AnnouncementView, error)
	FindAll(ctx context.Context) ([]AnnouncementView, error)
}

type AnnouncementRepository interface {
	AnnouncementQuery
	GetByID(ctx context.Context, id int) (*Announcement, error)
	Create(ctx context.Context, item *Announcement) error
	Update(ctx context.Context, item *Announcement) error
	Delete(ctx context.Context, id int) (bool, error)
}

func New(input SaveInput) (*Announcement, error) {
	item := &Announcement{}
	if err := item.Update(input); err != nil {
		return nil, err
	}
	return item, nil
}

func (a *Announcement) Update(input SaveInput) error {
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	if title == "" && body == "" {
		return ErrAnnouncementContentEmpty
	}
	if !input.ShowEnd.After(input.ShowStart) {
		return ErrAnnouncementInvalidRange
	}
	linkURL := strings.TrimSpace(input.LinkURL)
	if linkURL != "" {
		parsed, err := url.Parse(linkURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return ErrAnnouncementInvalidLink
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return ErrAnnouncementInvalidLink
		}
		linkURL = parsed.String()
	}

	a.Title = title
	a.Body = body
	a.Priority = input.Priority
	a.ShowStart = input.ShowStart
	a.ShowEnd = input.ShowEnd
	a.LinkURL = linkURL
	a.LinkTitle = strings.TrimSpace(input.LinkTitle)
	return nil
}
