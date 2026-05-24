package repository

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/course"
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
		LastSemester:  e.LastSemester,
		RatingCount:   e.RatingCount,
		RatingAvg:     e.RatingAvg,
		CreatedAt:     e.CreatedAt,
	}
}

type CourseRepository struct {
	db           *gorm.DB
	cache        *redis.Client
	searchConfig string
}

func NewCourseRepository(db *gorm.DB, cache ...*redis.Client) *CourseRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &CourseRepository{db: db, cache: client, searchConfig: SearchConfig(db)}
}

func (r *CourseRepository) baseCourseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&CourseEntity{}).Joins("MainTeacher")
}

func (r *CourseRepository) applyFilter(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	if len(f.CourseIDs) > 0 {
		db = db.Where("courses.id IN ?", f.CourseIDs)
	}
	if f.TeacherID > 0 {
		db = db.Where("courses.main_teacher_id = ?", f.TeacherID)
	}
	if f.ExcludeID > 0 {
		db = db.Where("courses.id != ?", f.ExcludeID)
	}
	if f.Q != "" {
		db = applySearchVectorFilter(db, r.searchConfig, "courses.search_vector", f.Q)
	}
	if f.Code != "" {
		db = db.Where("LOWER(courses.code) = LOWER(?)", f.Code)
	}
	if f.Name != "" {
		db = db.Where("courses.name = ?", f.Name)
	}
	if f.MainTeacherName != "" {
		db = db.Where(`"MainTeacher".name = ?`, f.MainTeacherName)
	}
	if f.Department != "" {
		db = db.Where("courses.department = ?", f.Department)
	}
	if f.Credit != nil {
		db = db.Where("courses.credit = ?", *f.Credit)
	}
	if f.HasReview != nil {
		if *f.HasReview {
			db = db.Where("courses.rating_count > 0")
		} else {
			db = db.Where("courses.rating_count = 0")
		}
	}
	if f.Language != "" {
		db = db.Where("courses.language = ?", f.Language)
	}
	if len(f.Categories) > 0 {
		db = db.Where("courses.categories && ?", pq.StringArray(f.Categories))
	}
	if len(f.TargetYears) > 0 {
		db = db.Where("courses.target_years && ?", pq.StringArray(f.TargetYears))
	}
	return db
}

func (r *CourseRepository) applySort(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	desc := !f.Ascend
	switch f.OrderBy {
	case "rating_count":
		return db.
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "rating_count"}, Desc: desc}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "rating_avg"}, Desc: true}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "code"}})
	case "rating_avg":
		return db.
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "rating_avg"}, Desc: true}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "rating_count"}, Desc: true}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "code"}})
	default:
		return db.
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "code"}}).
			Order(clause.OrderByColumn{Column: clause.Column{Table: "courses", Name: "id"}})
	}
}

func (r *CourseRepository) applyPagination(db *gorm.DB, f course.CourseFilter) *gorm.DB {
	if f.Page > 0 && f.PageSize > 0 {
		offset := (f.Page - 1) * f.PageSize
		db = db.Offset(offset).Limit(f.PageSize)
	}
	return db
}

func (r *CourseRepository) OfferedCourseExists(ctx context.Context, courseID int, semester string) (bool, error) {
	key := cacheKey("course", courseID, "offered", semester)
	if cached, ok := cacheGetJSON[bool](ctx, r.cache, key); ok {
		return *cached, nil
	}

	count, err := gorm.G[OfferedCourseEntity](r.db).Where("course_id = ? AND semester = ?", courseID, semester).Count(ctx, "id")
	if err != nil {
		return false, err
	}
	exists := count > 0
	cacheSetJSON(ctx, r.cache, key, exists)
	return exists, nil
}

func (r *CourseRepository) OfferedSemesterExists(ctx context.Context, semester string) (bool, error) {
	key := cacheKey("course", "offered", "semester", semester)
	if cached, ok := cacheGetJSON[bool](ctx, r.cache, key); ok {
		return *cached, nil
	}

	count, err := gorm.G[OfferedCourseEntity](r.db).Where("semester = ?", semester).Count(ctx, "id")
	if err != nil {
		return false, err
	}
	exists := count > 0
	cacheSetJSON(ctx, r.cache, key, exists)
	return exists, nil
}

func (r *CourseRepository) FindOfferedCourses(ctx context.Context, courseID int) ([]course.OfferedCourseView, error) {
	entities, err := gorm.G[OfferedCourseEntity](r.db).
		Where("course_id = ?", courseID).
		Order(clause.OrderByColumn{Column: clause.Column{Table: "offered_courses", Name: "semester"}, Desc: true}).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]course.OfferedCourseView, 0, len(entities))
	for _, e := range entities {
		oc := course.OfferedCourseView{
			ID:          e.ID,
			Semester:    e.Semester,
			Language:    e.Language,
			TargetYears: e.TargetYears,
			Categories:  e.Categories,
		}
		result = append(result, oc)
	}
	return result, nil
}

func (r *CourseRepository) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseView, int64, error) {
	db := r.baseCourseQuery(ctx)
	db = r.applyFilter(db, filter)

	var total int64
	countDB := r.baseCourseQuery(ctx)
	countDB = r.applyFilter(countDB, filter)
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = r.applySort(db, filter)
	db = r.applyPagination(db, filter)

	var entities []CourseEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	cs := make([]course.CourseView, len(entities))
	for i, e := range entities {
		cv := newCourseViewFromEntity(&e)
		cs[i] = *cv
	}
	return cs, total, nil
}

func (r *CourseRepository) Get(ctx context.Context, courseID int) (*course.Course, error) {
	key := cacheKey("course", courseID)
	if cached, ok := cacheGetJSON[course.Course](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[CourseEntity](r.db).Where("id = ?", courseID).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	c := newCourseDomain(&e)
	cacheSetJSON(ctx, r.cache, key, &c)
	return &c, nil
}

func (r *CourseRepository) GetDetail(ctx context.Context, courseID int) (*course.CourseDetailView, error) {
	key := cacheKey("course", courseID, "detail")
	if cached, ok := cacheGetJSON[course.CourseDetailView](ctx, r.cache, key); ok {
		return cached, nil
	}

	var entity CourseEntity
	err := r.db.WithContext(ctx).
		Joins("MainTeacher").
		Where("courses.id = ?", courseID).
		Take(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	result := newCourseDetailViewFromEntity(&entity)

	type ratingCount struct {
		Rating int
		Count  int
	}
	var dist []ratingCount
	if err := gorm.G[ReviewEntity](r.db).
		Select("rating, COUNT(*) AS count").
		Where("course_id = ?", courseID).
		Group("rating").
		Scan(ctx, &dist); err != nil {
		return nil, err
	}
	for _, d := range dist {
		if d.Rating >= 1 && d.Rating <= 5 {
			result.Rating.Distribution[d.Rating-1] = d.Count
		}
	}

	offeredCourses, err := r.FindOfferedCourses(ctx, courseID)
	if err != nil {
		return nil, err
	}
	result.OfferedCourses = offeredCourses

	cacheSetJSON(ctx, r.cache, key, result)
	return result, nil
}

func (r *CourseRepository) GetFilters(ctx context.Context) (*course.CourseFilters, error) {
	key := cacheKey("course", "filters")
	if cached, ok := cacheGetJSON[course.CourseFilters](ctx, r.cache, key); ok {
		return cached, nil
	}

	var credits []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&CourseEntity{}).
		Select("CAST(credit AS TEXT) AS name, COUNT(*) AS count").
		Group("credit").Order("credit").
		Scan(&credits).Error; err != nil {
		return nil, err
	}

	var departments []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&CourseEntity{}).
		Select("department AS name, COUNT(*) AS count").
		Group("department").Order("department").
		Scan(&departments).Error; err != nil {
		return nil, err
	}

	var categories []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&CourseEntity{}).
		Select("category AS name, COUNT(*) AS count").
		Joins("CROSS JOIN LATERAL unnest(courses.categories) AS category").
		Group("category").Order("category").
		Scan(&categories).Error; err != nil {
		return nil, err
	}

	var targetYears []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&CourseEntity{}).
		Select("year AS name, COUNT(*) AS count").
		Joins("CROSS JOIN LATERAL unnest(courses.target_years) AS year").
		Group("year").Order("year").
		Scan(&targetYears).Error; err != nil {
		return nil, err
	}

	var languages []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&CourseEntity{}).
		Select("language AS name, COUNT(*) AS count").
		Where("language != ''").
		Group("language").Order("language").
		Scan(&languages).Error; err != nil {
		return nil, err
	}

	var semesters []course.FilterItem
	if err := r.db.WithContext(ctx).Model(&OfferedCourseEntity{}).
		Select("semester AS name, COUNT(DISTINCT course_id) AS count").
		Where("semester != ''").
		Group("semester").Order("semester DESC").
		Scan(&semesters).Error; err != nil {
		return nil, err
	}

	filters := &course.CourseFilters{
		Credits:     credits,
		Departments: departments,
		Categories:  categories,
		TargetYears: targetYears,
		Languages:   languages,
		Semesters:   semesters,
	}
	cacheSetJSON(ctx, r.cache, key, filters)
	return filters, nil
}

var _ course.CourseRepository = (*CourseRepository)(nil)
var _ course.CourseQuery = (*CourseRepository)(nil)
