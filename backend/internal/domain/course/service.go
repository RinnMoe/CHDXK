package course

import (
	"context"

	"jcourse/internal/domain/auth"
	"jcourse/pkg/apperr"
)

var ErrUserCannotModerate = apperr.ErrUserCannotModerateReview

type Service struct {
	courseRepo CourseRepository
}

func NewService(courseRepo CourseRepository) *Service {
	return &Service{courseRepo: courseRepo}
}

type UpdateModeratorRemark struct {
	CourseID        int
	ModeratorRemark string
}

func (s *Service) UpdateModeratorRemark(ctx context.Context, u *auth.User, cmd UpdateModeratorRemark) error {
	if u == nil || !u.IsAdmin() {
		return ErrUserCannotModerate
	}
	c, err := s.courseRepo.Get(ctx, cmd.CourseID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCourseNotFound
	}
	return s.courseRepo.UpdateModeratorRemark(ctx, cmd.CourseID, cmd.ModeratorRemark)
}
