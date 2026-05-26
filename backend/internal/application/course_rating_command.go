package application

import (
	"context"

	"jcourse/internal/domain/course"
)

type CourseRatingCommandService struct {
	query  course.CourseQuery
	config course.RatingScoreConfig
}

func NewCourseRatingCommandService(query course.CourseQuery, config course.RatingScoreConfig) *CourseRatingCommandService {
	return &CourseRatingCommandService{query: query, config: config.Normalized()}
}

func (s *CourseRatingCommandService) RefreshRatingScores(ctx context.Context) error {
	return s.query.RefreshRatingScores(ctx, s.config)
}
