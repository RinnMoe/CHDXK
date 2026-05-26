package course

import "context"

type CourseRatingCommandService struct {
	repo   CourseRepository
	config RatingScoreConfig
}

func NewCourseRatingCommandService(repo CourseRepository, config RatingScoreConfig) *CourseRatingCommandService {
	return &CourseRatingCommandService{repo: repo, config: config.Normalized()}
}

func (s *CourseRatingCommandService) RefreshRatingScores(ctx context.Context) error {
	return s.repo.RefreshRatingScores(ctx, s.config)
}
