package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/teacher"
)

func newCourseDomain(e *CourseEntity) course.Course {
	return course.Course{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Credit:        e.Credit,
		MainTeacherID: e.MainTeacherID,
		CreatedAt:     e.CreatedAt,
	}
}

func newCourseQuery(e *courseRow) course.CourseForQuery {
	return course.CourseForQuery{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Credit:        e.Credit,
		Department:    e.Department,
		MainTeacherID: e.MainTeacherID,
		ReviewCount:   e.ReviewCount,
		AvgRating:     e.AvgRating,
		MainTeacher: &teacher.TeacherForQuery{
			ID:         e.TeacherID,
			Code:       e.TeacherCode,
			Name:       e.TeacherName,
			Department: e.TeacherDepartment,
			Title:      e.TeacherTitle,
		},
	}
}

type courseRow struct {
	ID            int
	Code          string
	Name          string
	Credit        float32
	Department    string
	MainTeacherID int
	ReviewCount   int
	AvgRating     float64

	TeacherID         int
	TeacherCode       string
	TeacherName       string
	TeacherDepartment string
	TeacherTitle      string
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) baseCourseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("courses c").
		Select(`c.id, c.code, c.name, c.credit, c.department, c.main_teacher_id,
			c.review_count, c.avg_rating,
			t.id AS teacher_id, t.code AS teacher_code, t.name AS teacher_name,
			t.department AS teacher_department, t.title AS teacher_title`).
		Joins("LEFT JOIN teachers t ON t.id = c.main_teacher_id")
}

func (r *CourseRepository) applyFilter(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	if f.TeacherID > 0 {
		db = db.Where("c.main_teacher_id = ?", f.TeacherID)
	}
	if f.ExcludeID > 0 {
		db = db.Where("c.id != ?", f.ExcludeID)
	}
	if f.Code != "" {
		db = db.Where("LOWER(c.code) = LOWER(?)", f.Code)
	}
	if f.Department != "" {
		db = db.Where("c.department = ?", f.Department)
	}
	if f.Credit != nil {
		db = db.Where("c.credit = ?", *f.Credit)
	}
	if f.HasReview != nil && *f.HasReview {
		db = db.Where("c.review_count > 0")
	}
	return db
}

func (r *CourseRepository) applySort(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	dir := "DESC"
	if f.OrderDir == "asc" {
		dir = "ASC"
	}
	switch f.OrderBy {
	case "review_count":
		db = db.Order(fmt.Sprintf("c.review_count %s", dir))
	case "avg_rating":
		db = db.Order(fmt.Sprintf("c.avg_rating %s", dir))
	default:
		db = db.Order("c.id DESC")
	}
	return db
}

func (r *CourseRepository) applyPagination(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	if f.Page > 0 && f.PageSize > 0 {
		offset := (f.Page - 1) * f.PageSize
		db = db.Offset(offset).Limit(f.PageSize)
	}
	return db
}

func (r *CourseRepository) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseForQuery, int64, error) {
	db := r.baseCourseQuery(ctx)
	db = r.applyFilter(db, filter)

	var total int64
	r.db.WithContext(ctx).Table("(?) AS sub", db).Count(&total)

	db = r.applySort(db, filter)
	db = r.applyPagination(db, filter)

	var rows []courseRow
	if err := db.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	cs := make([]course.CourseForQuery, len(rows))
	for i, row := range rows {
		cs[i] = newCourseQuery(&row)
	}
	return cs, total, nil
}

func (r *CourseRepository) Get(ctx context.Context, courseID int) (*course.Course, error) {
	e, err := gorm.G[CourseEntity](r.db).Where("id = ?", courseID).Take(ctx)
	if err != nil {
		return nil, err
	}
	return new(newCourseDomain(&e)), nil
}

func (r *CourseRepository) GetDetail(ctx context.Context, courseID int) (*course.CourseDetailForQuery, error) {
	var row courseRow
	err := r.baseCourseQuery(ctx).
		Where("c.id = ?", courseID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	result := &course.CourseDetailForQuery{
		CourseForQuery: newCourseQuery(&row),
	}

	type ratingCount struct {
		Rating int
		Count  int
	}
	var dist []ratingCount
	r.db.WithContext(ctx).Table("reviews").
		Select("rating, COUNT(*) AS count").
		Where("course_id = ? AND deleted_at IS NULL", courseID).
		Group("rating").
		Scan(&dist)
	for _, d := range dist {
		if d.Rating >= 1 && d.Rating <= 5 {
			result.RatingDistribution[d.Rating-1] = d.Count
		}
	}

	return result, nil
}

var _ course.CourseRepository = (*CourseRepository)(nil)
var _ course.CourseQuery = (*CourseRepository)(nil)
