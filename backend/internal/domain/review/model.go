package review

import (
	"context"
	"errors"
	"time"
)

type Review struct {
	ID        int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Score     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *Review) MakeRevision() Revision {
	return Revision{
		ReviewID:  r.ID,
		CourseID:  r.CourseID,
		Semester:  r.Semester,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		Score:     r.Score,
		CreatedAt: time.Now(),
	}
}

func (r *Review) ApplyUpdate(cmd Update) {
	r.Semester = cmd.Semester
	r.Rating = cmd.Rating
	r.Content = cmd.Content
	r.Score = cmd.Score
	r.UpdatedAt = cmd.Now
}

type Update struct {
	Semester string
	Rating   int
	Content  string
	Score    string
	Now      time.Time
}

func (r *Review) Validate() error {
	if r.Rating < 0 || r.Rating > 5 {
		return errors.New("invalid rating")
	}
	if r.Content == "" {
		return errors.New("content cannot be empty")
	}
	if len(r.Content) > 9681 {
		return errors.New("content cannot be longer than 9681 characters")
	}
	return nil
}

type Revision struct {
	ID        int
	ReviewID  int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Score     string
	CreatedAt time.Time
}

type ReviewRepository interface {
	Create(ctx context.Context, r *Review) error
	Update(ctx context.Context, r *Review, rv Revision) error
	Delete(ctx context.Context, r *Review) error
	Get(ctx context.Context, reviewID int) (*Review, error)
}
