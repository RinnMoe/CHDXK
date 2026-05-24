package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/course"
)

type CourseEnrollmentRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewCourseEnrollmentRepository(db *gorm.DB, cache ...*redis.Client) *CourseEnrollmentRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CourseEnrollmentRepository{db: db, cache: client}
}

func (r *CourseEnrollmentRepository) Create(ctx context.Context, enrollment *course.CourseEnrollment) error {
	entity := CourseEnrollmentEntity{
		UserID:    enrollment.UserID,
		CourseID:  enrollment.CourseID,
		Semester:  enrollment.Semester,
		CreatedAt: enrollment.CreatedAt,
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "course_id"}, {Name: "semester"}},
		DoNothing: true,
	}).Create(&entity).Error
	if err != nil {
		return err
	}
	enrollment.ID = entity.ID
	r.deleteEnrollmentCache(ctx, enrollment.UserID, enrollment.CourseID)
	return nil
}

func (r *CourseEnrollmentRepository) Delete(ctx context.Context, enrollmentID, userID int) error {
	var existing CourseEnrollmentEntity
	_ = r.db.WithContext(ctx).
		Select("course_id").
		Where("id = ? AND user_id = ?", enrollmentID, userID).
		Take(&existing).Error
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", enrollmentID, userID).
		Delete(&CourseEnrollmentEntity{}).Error; err != nil {
		return err
	}
	r.deleteEnrollmentCache(ctx, userID, existing.CourseID)
	return nil
}

func (r *CourseEnrollmentRepository) FindUserEnrollments(ctx context.Context, userID int) ([]course.CourseEnrollmentView, error) {
	key := cacheKey("course_enrollment", userID, "list")
	if cached, ok := cacheGetJSON[[]course.CourseEnrollmentView](ctx, r.cache, key); ok {
		return *cached, nil
	}
	views, err := r.findEnrollments(ctx, course.CourseEnrollmentFilter{UserID: userID})
	if err != nil {
		return nil, err
	}
	cacheSetJSON(ctx, r.cache, key, views)
	return views, nil
}

func (r *CourseEnrollmentRepository) FindUserCourseEnrollments(ctx context.Context, userID, courseID int) ([]course.CourseEnrollmentView, error) {
	key := cacheKey("course_enrollment", userID, "course", courseID)
	if cached, ok := cacheGetJSON[[]course.CourseEnrollmentView](ctx, r.cache, key); ok {
		return *cached, nil
	}
	views, err := r.findEnrollments(ctx, course.CourseEnrollmentFilter{UserID: userID, CourseID: courseID})
	if err != nil {
		return nil, err
	}
	cacheSetJSON(ctx, r.cache, key, views)
	return views, nil
}

func (r *CourseEnrollmentRepository) findEnrollments(ctx context.Context, filter course.CourseEnrollmentFilter) ([]course.CourseEnrollmentView, error) {
	db := r.db.WithContext(ctx).
		Model(&CourseEnrollmentEntity{}).
		Preload("Course.MainTeacher")

	db = applyEnrollmentFilter(db, filter)
	db = db.
		Order(clause.OrderByColumn{Column: clause.Column{Table: "course_enrollments", Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Table: "course_enrollments", Name: "id"}, Desc: true})

	var entities []CourseEnrollmentEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}

	views := make([]course.CourseEnrollmentView, 0, len(entities))
	for _, e := range entities {
		view := course.CourseEnrollmentView{
			ID:        e.ID,
			UserID:    e.UserID,
			Semester:  e.Semester,
			CreatedAt: e.CreatedAt,
		}
		if e.Course != nil {
			view.Course = *newCourseViewFromEntity(e.Course)
		}
		views = append(views, view)
	}
	return views, nil
}

func applyEnrollmentFilter(db *gorm.DB, filter course.CourseEnrollmentFilter) *gorm.DB {
	if filter.UserID > 0 {
		db = db.Where("course_enrollments.user_id = ?", filter.UserID)
	}
	if filter.CourseID > 0 {
		db = db.Where("course_enrollments.course_id = ?", filter.CourseID)
	}
	return db
}

func (r *CourseEnrollmentRepository) deleteEnrollmentCache(ctx context.Context, userID, courseID int) {
	keys := []string{cacheKey("course_enrollment", userID, "list")}
	if courseID > 0 {
		keys = append(keys, cacheKey("course_enrollment", userID, "course", courseID))
	}
	cacheDelete(ctx, r.cache, keys...)
}

var _ course.CourseEnrollmentRepository = (*CourseEnrollmentRepository)(nil)
var _ course.CourseEnrollmentQuery = (*CourseEnrollmentRepository)(nil)
