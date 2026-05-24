package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/teacher"
)

type TeacherRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewTeacherRepository(db *gorm.DB, cache ...*redis.Client) *TeacherRepository {
	var client *redis.Client
	if len(cache) > 0 {
		client = cache[0]
	}
	return &TeacherRepository{db: db, cache: client}
}

func newTeacherViewFromEntity(e *TeacherEntity) teacher.TeacherView {
	return teacher.TeacherView{
		ID:         e.ID,
		Code:       e.Code,
		Name:       e.Name,
		Department: e.Department,
		Title:      e.Title,
		Pinyin:     e.Pinyin,
		PinyinAbbr: e.PinyinAbbr,
	}
}

func (r *TeacherRepository) FindBy(ctx context.Context, filter teacher.TeacherFilter) ([]teacher.TeacherView, int64, error) {
	db := r.db.WithContext(ctx).Model(&TeacherEntity{})

	if len(filter.TeacherIDs) > 0 {
		db = db.Where("id IN ?", filter.TeacherIDs)
	}
	if filter.Department != "" {
		db = db.Where("department = ?", filter.Department)
	}
	if filter.Title != "" {
		db = db.Where("title = ?", filter.Title)
	}
	if filter.Q != "" {
		db = applySearchVectorFilter(db, "search_vector", filter.Q)
	}
	if filter.Code != "" {
		db = db.Where("LOWER(code) = LOWER(?)", filter.Code)
	}
	if filter.Name != "" {
		db = db.Where("name = ?", filter.Name)
	}
	if filter.Pinyin != "" {
		db = db.Where("pinyin = ?", filter.Pinyin)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "code"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}})

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		db = db.Offset(offset).Limit(filter.PageSize)
	}

	var entities []TeacherEntity
	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	result := make([]teacher.TeacherView, len(entities))
	for i, e := range entities {
		result[i] = newTeacherViewFromEntity(&e)
	}
	return result, total, nil
}

func (r *TeacherRepository) GetByID(ctx context.Context, teacherID int) (*teacher.TeacherView, error) {
	key := cacheKey("teacher", teacherID)
	if cached, ok := cacheGetJSON[teacher.TeacherView](ctx, r.cache, key); ok {
		return cached, nil
	}

	e, err := gorm.G[TeacherEntity](r.db).Where("id = ?", teacherID).Take(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	v := newTeacherViewFromEntity(&e)
	cacheSetJSON(ctx, r.cache, key, &v)
	return &v, nil
}

func (r *TeacherRepository) GetFilters(ctx context.Context) (*teacher.TeacherFilters, error) {
	key := cacheKey("teacher", "filters")
	if cached, ok := cacheGetJSON[teacher.TeacherFilters](ctx, r.cache, key); ok {
		return cached, nil
	}

	var departments []teacher.FilterItem
	if err := r.db.WithContext(ctx).Model(&TeacherEntity{}).
		Select("department AS name, COUNT(*) AS count").
		Group("department").Order("department").
		Scan(&departments).Error; err != nil {
		return nil, err
	}

	var titles []teacher.FilterItem
	if err := r.db.WithContext(ctx).Model(&TeacherEntity{}).
		Select("title AS name, COUNT(*) AS count").
		Group("title").Order("title").
		Scan(&titles).Error; err != nil {
		return nil, err
	}

	filters := &teacher.TeacherFilters{
		Departments: departments,
		Titles:      titles,
	}
	cacheSetJSON(ctx, r.cache, key, filters)
	return filters, nil
}

var _ teacher.TeacherQuery = (*TeacherRepository)(nil)
