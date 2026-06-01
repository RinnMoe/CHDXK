package repository

import (
	"fmt"
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

func applySearchVectorOrNameChainFilter[T any](db gorm.ChainInterface[T], config, vectorColumn, nameColumn, q string) gorm.ChainInterface[T] {
	query := searchQuery(q)
	if query == "" {
		return db
	}
	nameLike := "%" + query + "%"
	condition := fmt.Sprintf("(%s @@ websearch_to_tsquery(?::regconfig, ?) OR %s ILIKE ?)", vectorColumn, nameColumn)
	return db.Where(condition, config, query, nameLike)
}

func applyCourseSearchChainFilter[T any](db gorm.ChainInterface[T], config, q string) gorm.ChainInterface[T] {
	query := searchQuery(q)
	if query == "" {
		return db
	}
	nameLike := "%" + query + "%"
	codeLike := "%" + query + "%"
	return db.Where(
		`(courses.search_vector @@ websearch_to_tsquery(?::regconfig, ?)
			OR courses.code ILIKE ?
			OR courses.name ILIKE ?)`,
		config,
		query,
		codeLike,
		nameLike,
	)
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

func RefreshCourseSearchVectors(db *gorm.DB, semesters ...string) error {
	config := SearchConfig(db)
	query := `
		UPDATE courses AS c
		SET search_vector =
			setweight(to_tsvector(?::regconfig, coalesce(c.name, '')), 'A') ||
			setweight(to_tsvector('simple', coalesce(c.code, '')), 'A') ||
			setweight(to_tsvector(?::regconfig, coalesce(t.name, '')), 'B')
		FROM teachers AS t
		WHERE c.main_teacher_id = t.id
	`
	args := []any{config, config}
	if len(semesters) > 0 {
		query += "\n\t\tAND c.last_semester IN ?"
		args = append(args, semesters)
	}
	return db.Exec(query, args...).Error
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
