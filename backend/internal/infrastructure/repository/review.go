package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"jcourse/internal/domain/review"
)

type ReviewEntity struct {
	ID        int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Grade     string
	CreatedAt int64
	UpdatedAt int64
	DeletedAt int64
}

type ReviewRevisionEntity struct {
	ReviewID  int
	CourseID  int
	Semester  string
	UserID    int
	Rating    int
	Content   string
	Grade     string
	CreatedAt int64
}

func newReviewEntity(r *review.Review) ReviewEntity {
	return ReviewEntity{
		CourseID:  r.CourseID,
		Semester:  r.Semester,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		Grade:     r.Grade,
		CreatedAt: r.CreatedAt.Unix(),
		UpdatedAt: r.UpdatedAt.Unix(),
	}
}

func newReviewDomain(e *ReviewEntity) review.Review {
	return review.Review{
		CourseID:  e.CourseID,
		Semester:  e.Semester,
		UserID:    e.UserID,
		Rating:    e.Rating,
		Content:   e.Content,
		Grade:     e.Grade,
		CreatedAt: time.Unix(e.CreatedAt, 0),
		UpdatedAt: time.Unix(e.UpdatedAt, 0),
	}
}

func newReviewRevisionEntity(r review.Revision) ReviewRevisionEntity {
	return ReviewRevisionEntity{
		ReviewID:  r.ReviewID,
		CourseID:  r.CourseID,
		Semester:  r.Semester,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		Grade:     r.Grade,
		CreatedAt: r.CreatedAt.Unix(),
	}
}

func newReviewRevisionQuery(e *ReviewRevisionEntity) review.RevisionForQuery {
	return review.RevisionForQuery{
		Revision: review.Revision{
			ReviewID:  e.ReviewID,
			CourseID:  e.CourseID,
			Semester:  e.Semester,
			UserID:    e.UserID,
			Rating:    e.Rating,
			Content:   e.Content,
			Grade:     e.Grade,
			CreatedAt: time.Unix(e.CreatedAt, 0),
		},
	}
}

func newReviewQuery(e *ReviewEntity) review.ReviewForQuery {
	return review.ReviewForQuery{
		Review: review.Review{
			CourseID: e.CourseID,
			Semester: e.Semester,
			UserID:   e.UserID,
			Rating:   e.Rating,
			Content:  e.Content,
			Grade:    e.Grade,
		},
	}
}

type ReviewRepository struct {
	db *gorm.DB
}

func (r2 *ReviewRepository) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewForQuery, error) {
	db := gorm.G[ReviewEntity](r2.db).Where("deleted_at = 0")
	if filter.CourseID != 0 {
		db = db.Where("course_id = ?", filter.CourseID)
	}
	if filter.UserID != 0 {
		db = db.Where("user_id = ?", filter.UserID)
	}
	if filter.Semester != "" {
		db = db.Where("semester = ?", filter.Semester)
	}
	if filter.Rating != 0 {
		db = db.Where("rating = ?", filter.Rating)
	}
	if filter.Order != "" {
		db = db.Order(filter.Order)
	}
	es, err := db.Find(ctx)
	if err != nil {
		return nil, err
	}
	rs := make([]review.ReviewForQuery, len(es))
	for i, e := range es {
		rs[i] = newReviewQuery(&e)
	}
	return rs, nil
}

func (r2 *ReviewRepository) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionForQuery, error) {
	es, err := gorm.G[ReviewRevisionEntity](r2.db).Where("review_id = ? AND deleted_at = 0", reviewID).Find(ctx)
	if err != nil {
		return nil, err
	}
	rs := make([]review.RevisionForQuery, len(es))
	for i, e := range es {
		rs[i] = newReviewRevisionQuery(&e)
	}
	return rs, nil
}

func (r2 *ReviewRepository) Create(ctx context.Context, r *review.Review) error {
	e := newReviewEntity(r)
	if err := gorm.G[ReviewEntity](r2.db).Create(ctx, &e); err != nil {
		return err
	}
	r.ID = e.ID
	return nil
}

func (r2 *ReviewRepository) Update(ctx context.Context, r *review.Review) error {
	e := newReviewEntity(r)
	rr := newReviewRevisionEntity(r.MakeRevision())
	if err := r2.db.Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[ReviewEntity](tx).Where("id = ? AND deleted_at = 0", e.ID).Updates(ctx, e); err != nil {
			return err
		}
		if err := gorm.G[ReviewRevisionEntity](tx).Create(ctx, &rr); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	r.ID = e.ID
	return nil
}

func (r2 *ReviewRepository) Delete(ctx context.Context, reviewID int) error {
	if _, err := gorm.G[ReviewEntity](r2.db).Where("id = ? AND deleted_at = 0", reviewID).
		Update(ctx, "deleted_at", time.Now().Unix()); err != nil {
		return err
	}
	return nil
}

func (r2 *ReviewRepository) Get(ctx context.Context, reviewID int) (*review.Review, error) {
	e, err := gorm.G[ReviewEntity](r2.db).Where("id = ? AND deleted_at = 0", reviewID).First(ctx)
	if err != nil {
		return nil, err
	}
	r := newReviewDomain(&e)
	return &r, nil
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

var _ review.ReviewRepository = (*ReviewRepository)(nil)
var _ review.ReviewQuery = (*ReviewRepository)(nil)
