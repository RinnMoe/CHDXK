package application

import "jcourse/internal/domain/announcement"

type AnnouncementDTO struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Priority  int    `json:"priority"`
	CreatedAt string `json:"created_at"`
}

func newAnnouncementDTO(a *announcement.AnnouncementView) AnnouncementDTO {
	return AnnouncementDTO{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		Priority:  a.Priority,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
