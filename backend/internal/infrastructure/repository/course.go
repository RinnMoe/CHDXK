package repository

import (
	"context"
	"fmt"

	"github.com/lib/pq"
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
		Categories:    e.Categories,
		Language:      e.Language,
		TargetYears:   e.TargetYears,
		RatingCount:   e.ReviewCount,
		RatingAvg:     e.AvgRating,
		CreatedAt:     e.CreatedAt,
	}
}

func newCourseQuery(e *courseRow) course.CourseView {
	return course.CourseView{
		ID:            e.ID,
		Code:          e.Code,
		Name:          e.Name,
		Credit:        e.Credit,
		Department:    e.Department,
		MainTeacherID: e.MainTeacherID,
		Categories:    e.Categories,
		Language:      e.Language,
		TargetYears:   e.TargetYears,
		Rating: course.RatingInfo{
			Count: e.ReviewCount,
			Avg:   e.AvgRating,
		},
		MainTeacher: &teacher.TeacherView{
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
	Categories    pq.StringArray `gorm:"type:text[]"`
	Language      string
	TargetYears   pq.StringArray `gorm:"type:text[]"`
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
			c.categories, c.language, c.target_years,
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
	if f.Language != "" {
		db = db.Where("c.language = ?", f.Language)
	}
	if len(f.Categories) > 0 {
		db = db.Where("c.categories && ?", pq.StringArray(f.Categories))
	}
	if len(f.TargetYears) > 0 {
		db = db.Where("c.target_years && ?", pq.StringArray(f.TargetYears))
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

func (r *CourseRepository) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("offered_courses").
		Where("course_id = ? AND semester = ?", courseID, semester).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CourseRepository) FindOfferedCourses(ctx context.Context, courseID int) ([]course.OfferedCourseView, error) {
	type ocRow struct {
		ID          int
		Semester    string
		Language    string
		TargetYears pq.StringArray `gorm:"type:text[]"`
		Categories  pq.StringArray `gorm:"type:text[]"`
		TeacherID   int
		TeacherCode string
		TeacherName string
		TeacherDept string
		TeacherTitl string
	}

	var rows []ocRow
	err := r.db.WithContext(ctx).Table("offered_courses oc").
		Select("oc.id, oc.semester, oc.language, oc.target_years, oc.categories, "+
			"t.id AS teacher_id, t.code AS teacher_code, t.name AS teacher_name, "+
			"t.department AS teacher_dept, t.title AS teacher_titl").
		Joins("LEFT JOIN course_teacher_groups ctg ON ctg.offered_course_id = oc.id").
		Joins("LEFT JOIN teachers t ON t.id = ctg.teacher_id").
		Where("oc.course_id = ?", courseID).
		Order("oc.semester DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	ocMap := make(map[int]*course.OfferedCourseView)
	var ocOrder []int
	for _, row := range rows {
		oc, ok := ocMap[row.ID]
		if !ok {
			oc = &course.OfferedCourseView{
				ID:          row.ID,
				Semester:    row.Semester,
				Language:    row.Language,
				TargetYears: row.TargetYears,
				Categories:  row.Categories,
			}
			ocMap[row.ID] = oc
			ocOrder = append(ocOrder, row.ID)
		}
		if row.TeacherID > 0 {
			oc.TeacherGroup = append(oc.TeacherGroup, &teacher.TeacherView{
				ID:         row.TeacherID,
				Code:       row.TeacherCode,
				Name:       row.TeacherName,
				Department: row.TeacherDept,
				Title:      row.TeacherTitl,
			})
		}
	}

	result := make([]course.OfferedCourseView, 0, len(ocMap))
	for _, id := range ocOrder {
		result = append(result, *ocMap[id])
	}
	return result, nil
}

func (r *CourseRepository) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseView, int64, error) {
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

	cs := make([]course.CourseView, len(rows))
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

func (r *CourseRepository) GetDetail(ctx context.Context, courseID int) (*course.CourseDetailView, error) {
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

	result := &course.CourseDetailView{
		ID:            row.ID,
		Code:          row.Code,
		Name:          row.Name,
		Credit:        row.Credit,
		Department:    row.Department,
		MainTeacherID: row.MainTeacherID,
		Categories:    row.Categories,
		Language:      row.Language,
		TargetYears:   row.TargetYears,
		Rating: course.RatingInfo{
			Count: row.ReviewCount,
			Avg:   row.AvgRating,
		},
		MainTeacher: &teacher.TeacherView{
			ID:         row.TeacherID,
			Code:       row.TeacherCode,
			Name:       row.TeacherName,
			Department: row.TeacherDepartment,
			Title:      row.TeacherTitle,
		},
	}

	type ratingCount struct {
		Rating int
		Count  int
	}
	var dist []ratingCount
	r.db.WithContext(ctx).Table("reviews").
		Select("rating, COUNT(*) AS count").
		Where("course_id = ?", courseID).
		Group("rating").
		Scan(&dist)
	for _, d := range dist {
		if d.Rating >= 1 && d.Rating <= 5 {
			result.Rating.Distribution[d.Rating-1] = d.Count
		}
	}

	offeredCourses, err := r.FindOfferedCourses(ctx, courseID)
	if err != nil {
		return nil, err
	}
	result.OfferedCourses = make([]*course.OfferedCourseView, len(offeredCourses))
	for i := range offeredCourses {
		result.OfferedCourses[i] = &offeredCourses[i]
	}

	return result, nil
}

var _ course.CourseRepository = (*CourseRepository)(nil)
var _ course.CourseQuery = (*CourseRepository)(nil)
