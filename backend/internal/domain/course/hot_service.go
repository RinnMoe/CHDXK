package course

import (
	"context"
	"time"
)

type CourseHotService struct {
	repo              HotCourseRepository
	scores            HotScoreConfig
	hotCourseLocation *time.Location
}

func NewCourseHotService(repo HotCourseRepository, scores HotScoreConfig) *CourseHotService {
	loc, err := DefaultHotCourseLocation()
	if err != nil {
		panic(err)
	}
	return &CourseHotService{repo: repo, scores: scores, hotCourseLocation: loc}
}

func (s *CourseHotService) RecordActivity(ctx context.Context, payload RecordHotCourseActivityPayload) error {
	return s.RecordActivityWithScores(ctx, payload, s.scores)
}

func (s *CourseHotService) RecordActivityWithScores(ctx context.Context, payload RecordHotCourseActivityPayload, scores HotScoreConfig) error {
	if s.repo == nil {
		return nil
	}
	score, err := scores.ScoreForActivity(payload.Activity)
	if err != nil {
		return err
	}
	if score == 0 {
		return nil
	}
	now := time.Now()
	return s.repo.AddScore(ctx, payload.CourseID, score, CurrentHotCoursePeriods(now, s.hotCourseLocation)...)
}
