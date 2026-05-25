package application

import (
	"context"
	"time"

	"jcourse/internal/domain/course"
)

type CourseHotCommandService struct {
	repo              course.HotCourseRepository
	scores            course.HotScoreConfig
	hotCourseLocation *time.Location
}

func NewCourseHotCommandService(repo course.HotCourseRepository, scores course.HotScoreConfig) *CourseHotCommandService {
	loc, err := course.DefaultHotCourseLocation()
	if err != nil {
		panic(err)
	}
	return &CourseHotCommandService{repo: repo, scores: scores, hotCourseLocation: loc}
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
	now := time.Now()
	return s.repo.AddScore(ctx, payload.CourseID, score, course.CurrentHotCoursePeriods(now, s.hotCourseLocation)...)
}
