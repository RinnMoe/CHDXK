package repository

import (
	"context"

	"gorm.io/gorm"

	"jcourse/internal/domain/teacher"
)

type TeacherRepository struct {
	db *gorm.DB
}

func NewTeacherRepository(db *gorm.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

func (r *TeacherRepository) FindBy(ctx context.Context, filter teacher.TeacherFilter) ([]teacher.TeacherView, int64, error) {
	db := r.db.WithContext(ctx).Model(&TeacherEntity{})

	if filter.Department != "" {
		db = db.Where("department = ?", filter.Department)
	}
	if filter.Title != "" {
		db = db.Where("title = ?", filter.Title)
	}
	if filter.Pinyin != "" {
		like := "%" + filter.Pinyin + "%"
		db = db.Where("pinyin LIKE ? OR pinyin_abbr LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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
		result[i] = teacher.TeacherView{
			ID:         e.ID,
			Code:       e.Code,
			Name:       e.Name,
			Department: e.Department,
			Title:      e.Title,
			Pinyin:     e.Pinyin,
			PinyinAbbr: e.PinyinAbbr,
		}
	}
	return result, total, nil
}

func (r *TeacherRepository) GetFilters(ctx context.Context) (*teacher.TeacherFilters, error) {
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

	return &teacher.TeacherFilters{
		Departments: departments,
		Titles:      titles,
	}, nil
}

var _ teacher.TeacherQuery = (*TeacherRepository)(nil)
