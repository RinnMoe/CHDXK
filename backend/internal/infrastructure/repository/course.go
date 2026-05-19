package repository

import (
	"context"

	"gorm.io/gorm"

	"jcourse/internal/domain/course"
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

func newCourseQuery(e *CourseEntity) course.CourseForQuery {
	return course.CourseForQuery{
		ID:     e.ID,
		Code:   e.Code,
		Name:   e.Name,
		Credit: e.Credit,
	}
}

type CourseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (c *CourseRepository) FindBy(ctx context.Context, filter course.CourseFilter) ([]course.CourseForQuery, error) {
	db := gorm.G[CourseEntity](c.db).Where("deleted_at IS NULL")
	es, err := db.Find(ctx)
	if err != nil {
		return nil, err
	}

	cs := make([]course.CourseForQuery, len(es))
	for i, e := range es {
		cs[i] = newCourseQuery(&e)
	}

	return cs, nil
}

func (c *CourseRepository) Get(ctx context.Context, courseID int) (*course.Course, error) {
	e, err := gorm.G[CourseEntity](c.db).Where("id = ?", courseID).Take(ctx)
	if err != nil {
		return nil, err
	}
	return new(newCourseDomain(&e)), nil
}

var _ course.CourseRepository = (*CourseRepository)(nil)
var _ course.CourseQuery = (*CourseRepository)(nil)
