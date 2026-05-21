package application

import (
	"context"

	"jcourse/internal/domain/announcement"
)

type AnnouncementQueryService struct {
	announcementQuery announcement.AnnouncementQuery
}

func NewAnnouncementQueryService(announcementQuery announcement.AnnouncementQuery) *AnnouncementQueryService {
	return &AnnouncementQueryService{announcementQuery: announcementQuery}
}

func (s *AnnouncementQueryService) ListActiveAnnouncements(ctx context.Context) ([]AnnouncementDTO, error) {
	announcements, err := s.announcementQuery.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]AnnouncementDTO, len(announcements))
	for i, a := range announcements {
		result[i] = newAnnouncementDTO(&a)
	}
	return result, nil
}
