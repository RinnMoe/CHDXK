package application

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type CourseHotCommandService struct {
	repo   course.HotCourseRepository
	scores course.HotScoreConfig
}

func NewCourseHotCommandService(repo course.HotCourseRepository, scores course.HotScoreConfig) *CourseHotCommandService {
	return &CourseHotCommandService{repo: repo, scores: scores}
}

func (s *CourseHotCommandService) RecordActivity(ctx context.Context, payload course.RecordHotCourseActivityPayload) error {
	if s.repo == nil {
		return nil
	}
	score, err := s.scores.ScoreForActivity(payload.Activity)
	if err != nil {
		return err
	}
	if score == 0 {
		return nil
	}
	return s.repo.AddScore(ctx, payload.CourseID, score, time.Now())
}
