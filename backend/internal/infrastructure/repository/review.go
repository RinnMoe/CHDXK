package repository

import (
	"context"

	"gorm.io/gorm"

	"jcourse/internal/domain/review"
)

func newReviewEntity(r *review.Review) ReviewEntity {
	return ReviewEntity{
		ID:        r.ID,
		CourseID:  r.CourseID,
		Semester:  r.Semester,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Content:   r.Content,
		Score:     r.Score,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func newReviewDomain(e *ReviewEntity) review.Review {
	return review.Review{
		ID:        e.ID,
		CourseID:  e.CourseID,
		Semester:  e.Semester,
		UserID:    e.UserID,
		Rating:    e.Rating,
		Content:   e.Content,
		Score:     e.Score,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
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
		Score:     r.Score,
		CreatedAt: r.CreatedAt,
	}
}

func newReviewRevisionQuery(e *ReviewRevisionEntity) review.RevisionView {
	return review.RevisionView{
		ReviewID:  e.ReviewID,
		CourseID:  e.CourseID,
		Semester:  e.Semester,
		UserID:    e.UserID,
		Rating:    e.Rating,
		Content:   e.Content,
		Score:     e.Score,
		CreatedAt: e.CreatedAt,
	}
}

func newReviewQuery(e *ReviewEntity) review.ReviewView {
	return review.ReviewView{
		Review: review.Review{
			CourseID: e.CourseID,
			Semester: e.Semester,
			UserID:   e.UserID,
			Rating:   e.Rating,
			Content:  e.Content,
			Score:    e.Score,
		},
	}
}

type ReviewRepository struct {
	db *gorm.DB
}

func (r2 *ReviewRepository) updateCourseStats(tx *gorm.DB, courseID int) error {
	return tx.Exec(`UPDATE courses SET
		rating_count = (SELECT COUNT(*) FROM reviews WHERE course_id = ?),
		rating_avg = (SELECT COALESCE(AVG(rating), 0) FROM reviews WHERE course_id = ?)
		WHERE id = ?`, courseID, courseID, courseID).Error
}

func (r2 *ReviewRepository) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, error) {
	db := gorm.G[ReviewEntity](r2.db).Where("1 = 1")
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
	if !filter.CreatedAfter.IsZero() {
		db = db.Where("created_at > ?", filter.CreatedAfter)
	}
	if filter.Order != "" {
		db = db.Order(filter.Order)
	}
	if filter.Limit > 0 {
		db = db.Limit(filter.Limit)
	}
	es, err := db.Find(ctx)
	if err != nil {
		return nil, err
	}
	rs := make([]review.ReviewView, len(es))
	for i, e := range es {
		rs[i] = newReviewQuery(&e)
	}
	return rs, nil
}

func (r2 *ReviewRepository) FindRevisions(ctx context.Context, reviewID int) ([]review.RevisionView, error) {
	es, err := gorm.G[ReviewRevisionEntity](r2.db).Where("review_id = ?", reviewID).Find(ctx)
	if err != nil {
		return nil, err
	}
	rs := make([]review.RevisionView, len(es))
	for i, e := range es {
		rs[i] = newReviewRevisionQuery(&e)
	}
	return rs, nil
}

func (r2 *ReviewRepository) Create(ctx context.Context, r *review.Review) error {
	e := newReviewEntity(r)
	if err := r2.db.Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[ReviewEntity](tx).Create(ctx, &e); err != nil {
			return err
		}
		return r2.updateCourseStats(tx, r.CourseID)
	}); err != nil {
		return err
	}

	r.ID = e.ID
	return nil
}

func (r2 *ReviewRepository) Update(ctx context.Context, r *review.Review, rv review.Revision) error {
	e := newReviewEntity(r)
	if err := r2.db.Transaction(func(tx *gorm.DB) error {
		rr := newReviewRevisionEntity(rv)
		if _, err := gorm.G[ReviewEntity](tx).Where("id = ?", e.ID).Updates(ctx, e); err != nil {
			return err
		}
		if err := gorm.G[ReviewRevisionEntity](tx).Create(ctx, &rr); err != nil {
			return err
		}
		return r2.updateCourseStats(tx, r.CourseID)
	}); err != nil {
		return err
	}

	r.ID = e.ID
	return nil
}

func (r2 *ReviewRepository) Delete(ctx context.Context, r *review.Review) error {
	return r2.db.Transaction(func(tx *gorm.DB) error {
		if _, err := gorm.G[ReviewEntity](tx).Where("id = ?", r.ID).Delete(ctx); err != nil {
			return err
		}
		return r2.updateCourseStats(tx, r.CourseID)
	})
}

func (r2 *ReviewRepository) Get(ctx context.Context, reviewID int) (*review.Review, error) {
	e, err := gorm.G[ReviewEntity](r2.db).Where("id = ?", reviewID).First(ctx)
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
