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
	Grade     string
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
		Grade:     r.Grade,
		CreatedAt: time.Now(),
	}
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
	Grade     string
	CreatedAt time.Time
}

type ReviewRepository interface {
	Create(ctx context.Context, r *Review) error
	Update(ctx context.Context, r *Review) error
	Delete(ctx context.Context, reviewID int) error
	Get(ctx context.Context, reviewID int) (*Review, error)
}
