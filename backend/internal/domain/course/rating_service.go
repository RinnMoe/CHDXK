package course

import "context"

type CourseRatingCommandService struct {
	query  CourseQuery
	config RatingScoreConfig
}

func NewCourseRatingCommandService(query CourseQuery, config RatingScoreConfig) *CourseRatingCommandService {
	return &CourseRatingCommandService{query: query, config: config.Normalized()}
}

func (s *CourseRatingCommandService) RefreshRatingScores(ctx context.Context) error {
	return s.query.RefreshRatingScores(ctx, s.config)
}
