package application

import (
	"time"

	"jcourse/internal/domain/announcement"
)

type AnnouncementDTO struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Priority  int    `json:"priority"`
	ShowStart string `json:"show_start"`
	ShowEnd   string `json:"show_end"`
	LinkURL   string `json:"link_url"`
	LinkTitle string `json:"link_title"`
	CreatedAt string `json:"created_at"`
}

type SaveAnnouncementCommand struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Priority  int    `json:"priority"`
	ShowStart string `json:"show_start" binding:"required"`
	ShowEnd   string `json:"show_end" binding:"required"`
	LinkURL   string `json:"link_url"`
	LinkTitle string `json:"link_title"`
}

func newAnnouncementDTO(a *announcement.AnnouncementView) AnnouncementDTO {
	return AnnouncementDTO{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		Priority:  a.Priority,
		ShowStart: a.ShowStart.Format(time.RFC3339),
		ShowEnd:   a.ShowEnd.Format(time.RFC3339),
		LinkURL:   a.LinkURL,
		LinkTitle: a.LinkTitle,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}

func NewAnnouncementDTOFromDomain(a *announcement.Announcement) AnnouncementDTO {
	return AnnouncementDTO{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		Priority:  a.Priority,
		ShowStart: a.ShowStart.Format(time.RFC3339),
		ShowEnd:   a.ShowEnd.Format(time.RFC3339),
		LinkURL:   a.LinkURL,
		LinkTitle: a.LinkTitle,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
	}
}
