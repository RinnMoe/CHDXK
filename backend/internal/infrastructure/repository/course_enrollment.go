package repository

import (
	"context"
	"strings"
	"time"

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

func (r *CourseEnrollmentRepository) SyncFromCoursePairs(ctx context.Context, userID int, semester string, pairs []course.CourseCodeTeacher) (int64, error) {
	normalized := make([]course.CourseCodeTeacher, 0, len(pairs))
	seen := make(map[course.CourseCodeTeacher]struct{}, len(pairs))
	for _, pair := range pairs {
		pair.Code = strings.TrimSpace(pair.Code)
		pair.TeacherName = strings.TrimSpace(pair.TeacherName)
		if pair.Code == "" || pair.TeacherName == "" {
			continue
		}
		key := course.CourseCodeTeacher{Code: strings.ToLower(pair.Code), TeacherName: pair.TeacherName}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, pair)
	}
	if len(normalized) == 0 {
		return 0, nil
	}

	db := r.db.WithContext(ctx).
		Model(&CourseEntity{}).
		Joins("MainTeacher").
		Distinct("courses.id")
	for i, pair := range normalized {
		condition := `LOWER(courses.code) = LOWER(?) AND "MainTeacher".name = ?`
		if i == 0 {
			db = db.Where(condition, pair.Code, pair.TeacherName)
		} else {
			db = db.Or(condition, pair.Code, pair.TeacherName)
		}
	}

	var matched []CourseEntity
	if err := db.Find(&matched).Error; err != nil {
		return 0, err
	}
	if len(matched) == 0 {
		return 0, nil
	}

	now := time.Now()
	entities := make([]CourseEnrollmentEntity, 0, len(matched))
	courseIDs := make([]int, 0, len(matched))
	for _, c := range matched {
		entities = append(entities, CourseEnrollmentEntity{
			UserID:    userID,
			CourseID:  c.ID,
			Semester:  semester,
			CreatedAt: now,
		})
		courseIDs = append(courseIDs, c.ID)
	}

	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "course_id"}, {Name: "semester"}},
		DoUpdates: clause.Assignments(map[string]any{"semester": semester}),
	}).Create(&entities)
	if result.Error != nil {
		return 0, result.Error
	}
	r.deleteEnrollmentCache(ctx, userID, courseIDs...)
	return result.RowsAffected, nil
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

func (r *CourseEnrollmentRepository) deleteEnrollmentCache(ctx context.Context, userID int, courseIDs ...int) {
	keys := []string{cacheKey("course_enrollment", userID, "list")}
	for _, courseID := range courseIDs {
		if courseID <= 0 {
			continue
		}
		keys = append(keys, cacheKey("course_enrollment", userID, "course", courseID))
	}
	cacheDelete(ctx, r.cache, keys...)
}

var _ course.CourseEnrollmentRepository = (*CourseEnrollmentRepository)(nil)
var _ course.CourseEnrollmentQuery = (*CourseEnrollmentRepository)(nil)
