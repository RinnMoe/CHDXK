package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/course"
	"jcourse/pkg/apperr"
)

type CourseEnrollmentRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func (r *CourseEnrollmentRepository) Save(ctx context.Context, enrollmentID, userID, sectionID int) error {
	enrollment := &course.CourseEnrollment{ID: enrollmentID, UserID: userID, CreatedAt: time.Now()}
	previousCourseID := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if enrollment.ID > 0 {
			var existing CourseEnrollmentEntity
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", enrollment.ID, enrollment.UserID).Take(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperr.NotFound("选课记录不存在")
				}
				return err
			}
			previousCourseID = existing.CourseID
		}
		var selection struct {
			CourseID int
			Semester string
		}
		if err := tx.Table("course_sections AS section").Select("offering.course_id, offering.semester").Joins("JOIN offered_courses AS offering ON offering.id = section.offered_course_id").Where("section.id = ?", sectionID).Take(&selection).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperr.BadRequest("该教学班已不存在，请重新选择")
			}
			return err
		}
		enrollment.CourseID, enrollment.Semester = selection.CourseID, selection.Semester
		var count int64
		if err := tx.Model(&CourseEntity{}).Where("id = ?", enrollment.CourseID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return course.ErrCourseNotFound
		}
		if err := tx.Model(&CourseEnrollmentEntity{}).Where("user_id = ? AND course_id = ? AND semester = ? AND id <> ?", enrollment.UserID, enrollment.CourseID, enrollment.Semester, enrollment.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return apperr.BadRequest("该课程在此学期已有选课记录")
		}
		if enrollment.ID == 0 {
			entity := CourseEnrollmentEntity{UserID: enrollment.UserID, CourseID: enrollment.CourseID, Semester: enrollment.Semester, CreatedAt: enrollment.CreatedAt}
			if err := tx.Create(&entity).Error; err != nil {
				return err
			}
			enrollment.ID = entity.ID
			return nil
		}
		return tx.Model(&CourseEnrollmentEntity{}).Where("id = ? AND user_id = ?", enrollment.ID, enrollment.UserID).Updates(map[string]any{"course_id": enrollment.CourseID, "semester": enrollment.Semester}).Error
	})
	var sqlError interface{ SQLState() string }
	if errors.As(err, &sqlError) && sqlError.SQLState() == "23505" {
		return apperr.BadRequest("该课程在此学期已有选课记录")
	}
	if err == nil {
		r.deleteEnrollmentCache(ctx, enrollment.UserID, previousCourseID, enrollment.CourseID)
	}
	return err
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
