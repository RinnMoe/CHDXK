package repository

import (
	"context"
	"fmt"

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

func newReviewView(e *ReviewEntity) review.ReviewView {
	v := review.ReviewView{
		ID:       e.ID,
		CourseID: e.CourseID,
		Semester: e.Semester,
		UserID:   e.UserID,
		Rating:   e.Rating,
		Content:  e.Content,
		Score:    e.Score,
		Vote: review.ReviewVoteStats{
			LikeCount:    e.LikeCount,
			DislikeCount: e.DislikeCount,
		},
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
	if e.Course != nil {
		v.Course = newCourseViewFromEntity(e.Course)
	}
	return v
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

func (r2 *ReviewRepository) FindBy(ctx context.Context, filter review.ReviewFilter) ([]review.ReviewView, int64, error) {
	db := r2.db.WithContext(ctx).Model(&ReviewEntity{})

	if filter.WithCourse {
		db = db.Joins("Course").Joins("Course.MainTeacher")
	}

	db = r2.applyFilter(db, filter)

	var total int64
	db.Count(&total)

	db = r2.applySort(db, filter)
	db = r2.applyPagination(db, filter)

	var entities []ReviewEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	rs := make([]review.ReviewView, len(entities))
	for i, e := range entities {
		rs[i] = newReviewView(&e)
	}
	return rs, total, nil
}

func (r2 *ReviewRepository) applyFilter(db *gorm.DB, filter review.ReviewFilter) *gorm.DB {
	if filter.ReviewID != 0 {
		db = db.Where("id = ?", filter.ReviewID)
	}
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
	return db
}

func (r2 *ReviewRepository) applySort(db *gorm.DB, filter review.ReviewFilter) *gorm.DB {
	dir := "DESC"
	if filter.OrderDir == "asc" {
		dir = "ASC"
	}
	switch filter.OrderBy {
	case "rating":
		db = db.Order(fmt.Sprintf("rating %s", dir))
	case "like_count":
		db = db.Order(fmt.Sprintf("like_count %s", dir))
	case "created_at":
		db = db.Order(fmt.Sprintf("created_at %s", dir))
	default:
		db = db.Order(fmt.Sprintf("id %s", dir))
	}
	return db
}

func (r2 *ReviewRepository) applyPagination(db *gorm.DB, filter review.ReviewFilter) *gorm.DB {
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		db = db.Offset(offset).Limit(filter.PageSize)
	}
	return db
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
