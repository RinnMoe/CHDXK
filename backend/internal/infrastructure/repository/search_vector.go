package repository

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const preferredChineseSearchConfig = "jiebacfg"
const fallbackSearchConfig = "simple"

func SearchConfig(db *gorm.DB) string {
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

func applySearchVectorFilter(db *gorm.DB, config, column, q string) *gorm.DB {
	query := searchQuery(q)
	if query == "" {
		return db
	}
	return db.Where(column+" @@ websearch_to_tsquery(?::regconfig, ?)", config, query)
}

func searchRankOrder(column string, q string) string {
	query := searchQuery(q)
	if query == "" {
		return ""
	}
	return "ts_rank(" + column + ", websearch_to_tsquery(?::regconfig, ?)) DESC"
}

func TeacherSearchVectorExpr(config, code, searchName string) clause.Expr {
	return clause.Expr{
		SQL: `
			setweight(to_tsvector('simple', coalesce(?, '')), 'A') ||
			setweight(to_tsvector(?::regconfig, coalesce(?, '')), 'A')
		`,
		Vars: []any{code, config, searchName},
	}
}

func RefreshCourseSearchVectors(db *gorm.DB) error {
	config := SearchConfig(db)
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
	config := SearchConfig(db)
	return db.Exec(`
		UPDATE reviews
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(content, '')), 'A')
	`, config).Error
}

func refreshReviewSearchVector(tx *gorm.DB, config string, reviewID int) error {
	return tx.Exec(`
		UPDATE reviews
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(content, '')), 'A')
		WHERE id = ?
	`, config, reviewID).Error
}
