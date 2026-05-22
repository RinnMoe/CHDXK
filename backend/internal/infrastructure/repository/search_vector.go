package repository

import (
	"strings"

	"gorm.io/gorm"
)

const preferredChineseSearchConfig = "jiebacfg"
const fallbackSearchConfig = "simple"

func searchConfig(db *gorm.DB) string {
	var exists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_ts_config WHERE cfgname = ?)", preferredChineseSearchConfig).Scan(&exists).Error
	if err == nil && exists {
		return preferredChineseSearchConfig
	}
	return fallbackSearchConfig
}

func searchQuery(q string) string {
	return strings.TrimSpace(q)
}

func applySearchVectorFilter(db *gorm.DB, column, q string) *gorm.DB {
	query := searchQuery(q)
	if query == "" {
		return db
	}
	return db.Where(column+" @@ websearch_to_tsquery(?::regconfig, ?)", searchConfig(db), query)
}

func searchRankOrder(column string, q string) string {
	query := searchQuery(q)
	if query == "" {
		return ""
	}
	return "ts_rank(" + column + ", websearch_to_tsquery(?::regconfig, ?)) DESC"
}

func RefreshTeacherSearchVectors(db *gorm.DB) error {
	config := searchConfig(db)
	return db.Exec(`
		UPDATE teachers
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(code, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(pinyin, '')), 'B') ||
			setweight(to_tsvector('simple', coalesce(pinyin_abbr, '')), 'B')
	`, config).Error
}

func RefreshCourseSearchVectors(db *gorm.DB) error {
	config := searchConfig(db)
	return db.Exec(`
		UPDATE courses AS c
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(c.name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(c.code, '')), 'A') ||
			setweight(to_tsvector(?::regconfig, coalesce(t.name, '')), 'B')
		FROM teachers AS t
		WHERE c.main_teacher_id = t.id
	`, config, config).Error
}

func RefreshReviewSearchVectors(db *gorm.DB) error {
	config := searchConfig(db)
	return db.Exec(`
		UPDATE reviews
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(content, '')), 'A') ||
			setweight(to_tsvector(?::regconfig, coalesce(score, '')), 'B')
	`, config, config).Error
}

func refreshReviewSearchVector(tx *gorm.DB, reviewID int) error {
	config := searchConfig(tx)
	return tx.Exec(`
		UPDATE reviews
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(content, '')), 'A') ||
			setweight(to_tsvector(?::regconfig, coalesce(score, '')), 'B')
		WHERE id = ?
	`, config, config, reviewID).Error
}
