package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	pinyin "github.com/mozillazg/go-pinyin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jcourse/internal/domain/auth"
	"jcourse/internal/infrastructure/repository"
)

const batchSize = 1000

type Migrator struct {
	source     *gorm.DB
	target     *gorm.DB
	checkpoint *migrationCheckpoint
}

func NewMigrator(source, target *gorm.DB, checkpointPath string) *Migrator {
	checkpoint, err := loadCheckpoint(checkpointPath)
	if err != nil {
		panic(err)
	}
	return &Migrator{source: source, target: target, checkpoint: checkpoint}
}

func (m *Migrator) Run() error {
	log.Println("Starting v1 data migration...")
	ctx := &migrationContext{target: m.target}

	if err := m.migrateTeachers(ctx); err != nil {
		return fmt.Errorf("migrate teachers: %w", err)
	}
	if err := m.migrateCourses(ctx); err != nil {
		return fmt.Errorf("migrate courses: %w", err)
	}
	if err := m.migrateUsers(ctx); err != nil {
		return fmt.Errorf("migrate users: %w", err)
	}
	if err := m.migrateUserPoints(ctx); err != nil {
		return fmt.Errorf("migrate user points: %w", err)
	}
	if err := m.migrateReviews(ctx); err != nil {
		return fmt.Errorf("migrate reviews: %w", err)
	}
	if err := m.migrateReviewRevisions(ctx); err != nil {
		return fmt.Errorf("migrate review revisions: %w", err)
	}
	if err := m.migrateVotes(ctx); err != nil {
		return fmt.Errorf("migrate votes: %w", err)
	}
	if err := m.migrateCourseNotifications(ctx); err != nil {
		return fmt.Errorf("migrate course notifications: %w", err)
	}
	if err := m.refreshDerivedData(); err != nil {
		return fmt.Errorf("refresh derived data: %w", err)
	}
	if err := m.resetSequences(); err != nil {
		return fmt.Errorf("reset sequences: %w", err)
	}

	log.Println("V1 data migration complete.")
	return nil
}

type migrationContext struct {
	target *gorm.DB
}

func (m *Migrator) migrateTeachers(ctx *migrationContext) error {
	const stage = "teachers"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating teachers... already done, skipping")
		return nil
	}
	log.Println("Migrating teachers...")

	now := time.Now()
	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming teachers after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyTeachers(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.TeacherEntity, 0, len(rows))
		for _, row := range rows {
			code := strings.TrimSpace(nullStringValue(row.TID))
			if code == "" {
				return fmt.Errorf("legacy teacher %d has empty tid", row.ID)
			}
			fullPy, abbrPy := generatePinyin(row.Name)
			batch = append(batch, repository.TeacherEntity{
				ID:           row.ID,
				Code:         code,
				Name:         row.Name,
				Department:   legacyDepartmentName(row.Department),
				Title:        nullStringValue(row.Title),
				Pinyin:       fullPy,
				PinyinAbbr:   abbrPy,
				LastSemester: legacySemesterName(row.LastSemester),
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		}
		if err := upsertTeachers(ctx.target, batch); err != nil {
			return err
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Teachers: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Teachers: %d migrated", total)
	return nil
}

func (m *Migrator) migrateCourses(ctx *migrationContext) error {
	const stage = "courses"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating courses... already done, skipping")
		return nil
	}
	log.Println("Migrating courses...")

	now := time.Now()
	total := 0
	skipped := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming courses after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyCourses(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.CourseEntity, 0, len(rows))
		for _, row := range rows {
			if row.MainTeacherID == 0 {
				skipped++
				continue
			}
			courseTeachers := legacyTeacherIDs(row.TeacherGroup)
			if len(courseTeachers) == 0 {
				courseTeachers = []int64{int64(row.MainTeacherID)}
			}
			batch = append(batch, repository.CourseEntity{
				ID:            row.ID,
				Code:          row.Code,
				Name:          row.Name,
				Credit:        row.Credit,
				Department:    legacyDepartmentName(row.Department),
				MainTeacherID: row.MainTeacherID,
				Categories:    pq.StringArray(legacyCategoryNames(row.Categories)),
				Language:      "",
				TargetYears:   pq.StringArray{},
				TeacherIDs:    pq.Int64Array(courseTeachers),
				LastSemester:  legacySemesterName(row.LastSemester),
				RatingCount:   intOrZero(row.ReviewCount),
				RatingAvg:     floatOrZero(row.ReviewAvg),
				CreatedAt:     now,
			})
		}
		if len(batch) > 0 {
			if err := upsertCourses(ctx.target, batch); err != nil {
				return err
			}
		}
		total += len(batch)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Courses: %d migrated, %d skipped...", total, skipped)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Courses: %d migrated, %d skipped", total, skipped)
	return nil
}

func (m *Migrator) migrateUsers(ctx *migrationContext) error {
	const stage = "users"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating users... already done, skipping")
		return nil
	}
	log.Println("Migrating users...")

	now := time.Now()
	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming users after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyUsers(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.UserEntity, 0, len(rows))
		for _, row := range rows {
			createdAt := row.DateJoined
			if createdAt.IsZero() {
				createdAt = now
			}
			batch = append(batch, repository.UserEntity{
				ID:           row.ID,
				Username:     row.Username,
				Email:        sql.NullString{},
				Role:         userRole(row),
				PasswordHash: migratePasswordHash(row.PasswordHash),
				CreatedAt:    createdAt,
				LastSeenAt:   userLastSeen(row, createdAt),
				SuspendedAt:  nil,
				SuspendTill:  nil,
			})
		}
		if err := upsertUsers(ctx.target, batch); err != nil {
			return err
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Users: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Users: %d migrated", total)
	return nil
}

func (m *Migrator) migrateReviews(ctx *migrationContext) error {
	const stage = "reviews"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating reviews... already done, skipping")
		return nil
	}
	log.Println("Migrating reviews...")
	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming reviews after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyReviews(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.ReviewEntity, 0, len(rows))
		for _, row := range rows {
			updatedAt := row.CreatedAt
			if row.ModifiedAt.Valid {
				updatedAt = row.ModifiedAt.Time
			}
			batch = append(batch, repository.ReviewEntity{
				ID:              row.ID,
				CourseID:        row.CourseID,
				Semester:        nullStringValue(row.SemesterName),
				UserID:          row.UserID,
				Rating:          row.Rating,
				Content:         row.Comment,
				Score:           nullStringValue(row.Score),
				ModeratorRemark: nullStringValue(row.ModeratorRemark),
				LikeCount:       intOrZero(row.ApproveCount),
				DislikeCount:    intOrZero(row.DisapproveCount),
				CreatedAt:       row.CreatedAt,
				UpdatedAt:       updatedAt,
			})
		}
		if err := upsertReviews(ctx.target, batch); err != nil {
			return err
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Reviews: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Reviews: %d migrated", total)
	return nil
}

func (m *Migrator) migrateUserPoints(ctx *migrationContext) error {
	const stage = "user_points"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating user points... already done, skipping")
		return nil
	}
	log.Println("Migrating user points...")

	now := time.Now()
	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming user points after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyUserPoints(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.UserPointRecordEntity, 0, len(rows))
		for _, row := range rows {
			description := strings.TrimSpace(nullStringValue(row.Description))
			if description == "" {
				description = "legacy_migration"
			}
			createdAt := row.Time
			if createdAt.IsZero() {
				createdAt = now
			}
			batch = append(batch, repository.UserPointRecordEntity{
				ID:          row.ID,
				UserID:      row.UserID,
				Reason:      "legacy_migration",
				Amount:      row.Value,
				Description: description,
				CreatedAt:   createdAt,
			})
		}
		if err := upsertUserPoints(ctx.target, batch); err != nil {
			return err
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  User points: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  User points: %d migrated", total)
	return nil
}

func (m *Migrator) migrateVotes(ctx *migrationContext) error {
	const stage = "review_votes"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating review votes... already done, skipping")
		return nil
	}
	log.Println("Migrating review votes...")

	now := time.Now()
	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming review votes after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyReactions(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.ReviewVoteEntity, 0, len(rows))
		for _, row := range rows {
			updatedAt := now
			if row.ModifiedAt.Valid {
				updatedAt = row.ModifiedAt.Time
			}
			batch = append(batch, repository.ReviewVoteEntity{
				ReviewID:  row.ReviewID,
				UserID:    row.UserID,
				VoteType:  row.Reaction,
				CreatedAt: updatedAt,
				UpdatedAt: updatedAt,
			})
		}
		if err := upsertVotes(ctx.target, batch); err != nil {
			return err
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Review votes: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Review votes: %d migrated", total)
	return nil
}

func (m *Migrator) migrateReviewRevisions(ctx *migrationContext) error {
	const stage = "review_revisions"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating review revisions... already done, skipping")
		return nil
	}
	log.Println("Migrating review revisions...")

	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming review revisions after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyReviewRevisions(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.ReviewRevisionEntity, 0, len(rows))
		for _, row := range rows {
			if !row.ReviewID.Valid || !row.CourseID.Valid || !row.UserID.Valid {
				continue
			}
			batch = append(batch, repository.ReviewRevisionEntity{
				ID:        row.ID,
				ReviewID:  int(row.ReviewID.Int64),
				CourseID:  int(row.CourseID.Int64),
				Semester:  nullStringValue(row.SemesterName),
				UserID:    int(row.UserID.Int64),
				Rating:    row.Rating,
				Content:   row.Comment,
				Score:     nullStringValue(row.Score),
				CreatedAt: row.CreatedAt,
			})
		}
		if len(batch) > 0 {
			if err := upsertReviewRevisions(ctx.target, batch); err != nil {
				return err
			}
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Review revisions: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Review revisions: %d migrated", total)
	return nil
}

func (m *Migrator) migrateCourseNotifications(ctx *migrationContext) error {
	const stage = "course_notifications"
	if m.checkpoint.StageDone(stage) {
		log.Println("Migrating course notifications... already done, skipping")
		return nil
	}
	log.Println("Migrating course notifications...")

	total := 0
	lastID := m.checkpoint.StageLastID(stage)
	if lastID > 0 {
		log.Printf("  Resuming course notifications after legacy id %d", lastID)
	}
	for {
		rows, err := queryLegacyCourseNotifications(m.source, lastID)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}

		batch := make([]repository.CourseNotificationEntity, 0, len(rows))
		for _, row := range rows {
			if !row.UserID.Valid || !row.CourseID.Valid {
				continue
			}
			batch = append(batch, repository.CourseNotificationEntity{
				UserID:    int(row.UserID.Int64),
				CourseID:  int(row.CourseID.Int64),
				Level:     row.NotificationLevel,
				CreatedAt: row.ModifiedAt,
				UpdatedAt: row.ModifiedAt,
			})
		}
		if len(batch) > 0 {
			if err := upsertCourseNotifications(ctx.target, batch); err != nil {
				return err
			}
		}
		total += len(rows)
		lastID = rows[len(rows)-1].ID
		if err := m.checkpoint.MarkProgress(stage, lastID); err != nil {
			return err
		}
		log.Printf("  Course notifications: %d migrated...", total)
	}
	if err := m.checkpoint.MarkDone(stage); err != nil {
		return err
	}
	log.Printf("  Course notifications: %d migrated", total)
	return nil
}

func queryLegacyTeachers(db *gorm.DB, lastID int) ([]legacyTeacher, error) {
	var rows []legacyTeacher
	err := db.
		Preload("Department").
		Preload("LastSemester").
		Where("id > ?", lastID).
		Order("id").
		Limit(batchSize).
		Find(&rows).Error
	return rows, err
}

func queryLegacyCourses(db *gorm.DB, lastID int) ([]legacyCourse, error) {
	var rows []legacyCourse
	err := db.
		Preload("Department").
		Preload("LastSemester").
		Preload("Categories").
		Preload("TeacherGroup").
		Where("id > ?", lastID).
		Order("id").
		Limit(batchSize).
		Find(&rows).Error
	return rows, err
}

func queryLegacyUsers(db *gorm.DB, lastID int) ([]legacyUser, error) {
	var rows []legacyUser
	err := db.Raw(`
		SELECT u.id,
		       u.password AS password_hash,
		       u.username,
		       u.is_staff,
		       u.is_superuser,
		       u.date_joined,
		       u.last_login,
		       p.last_seen_at
		FROM auth_user AS u
		LEFT JOIN oauth_userprofile AS p ON p.user_id = u.id
		WHERE u.id > ?
		ORDER BY u.id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func queryLegacyReviews(db *gorm.DB, lastID int) ([]legacyReview, error) {
	var rows []legacyReview
	err := db.Raw(`
		SELECT r.id,
		       r.user_id,
		       r.course_id,
		       s.name AS semester_name,
		       r.rating,
		       r.comment,
		       r.created_at,
		       r.modified_at,
		       r.score,
		       r.moderator_remark,
		       r.approve_count,
		       r.disapprove_count
		FROM jcourse_api_review AS r
		LEFT JOIN jcourse_api_semester AS s ON s.id = r.semester_id
		WHERE r.id > ?
		ORDER BY r.id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func queryLegacyReviewRevisions(db *gorm.DB, lastID int) ([]legacyReviewRevision, error) {
	var rows []legacyReviewRevision
	err := db.Raw(`
		SELECT rr.id,
		       rr.review_id,
		       rr.course_id,
		       s.name AS semester_name,
		       rr.user_id,
		       rr.rating,
		       rr.comment,
		       rr.created_at,
		       rr.score
		FROM jcourse_api_reviewrevision AS rr
		LEFT JOIN jcourse_api_semester AS s ON s.id = rr.semester_id
		WHERE rr.id > ?
		ORDER BY rr.id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func queryLegacyReactions(db *gorm.DB, lastID int) ([]legacyReaction, error) {
	var rows []legacyReaction
	err := db.Raw(`
		SELECT id, user_id, review_id, reaction, modified_at
		FROM jcourse_api_reviewreaction
		WHERE id > ? AND reaction IN (1, -1)
		ORDER BY id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func queryLegacyUserPoints(db *gorm.DB, lastID int) ([]legacyUserPoint, error) {
	var rows []legacyUserPoint
	err := db.Raw(`
		SELECT id, user_id, value, description, time
		FROM jcourse_api_userpoint
		WHERE id > ?
		ORDER BY id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func queryLegacyCourseNotifications(db *gorm.DB, lastID int) ([]legacyCourseNotificationLevel, error) {
	var rows []legacyCourseNotificationLevel
	err := db.Raw(`
		SELECT id, user_id, course_id, notification_level, modified_at
		FROM jcourse_api_coursenotificationlevel
		WHERE id > ? AND notification_level IN (1, 2)
		ORDER BY id
		LIMIT ?
	`, lastID, batchSize).Scan(&rows).Error
	return rows, err
}

func upsertTeachers(db *gorm.DB, batch []repository.TeacherEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertCourses(db *gorm.DB, batch []repository.CourseEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertUsers(db *gorm.DB, batch []repository.UserEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertReviews(db *gorm.DB, batch []repository.ReviewEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertReviewRevisions(db *gorm.DB, batch []repository.ReviewRevisionEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertVotes(db *gorm.DB, batch []repository.ReviewVoteEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "review_id"}, {Name: "user_id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertUserPoints(db *gorm.DB, batch []repository.UserPointRecordEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func upsertCourseNotifications(db *gorm.DB, batch []repository.CourseNotificationEntity) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "course_id"}},
		DoNothing: true,
	}).Create(&batch).Error
}

func (m *Migrator) refreshDerivedData() error {
	const stage = "refresh_derived_data"
	if m.checkpoint.StageDone(stage) {
		log.Println("Refreshing search vectors and denormalized counters... already done, skipping")
		return nil
	}
	db := m.target
	log.Println("Refreshing search vectors and denormalized counters...")
	if err := db.Exec(`
		UPDATE reviews AS r SET
			like_count = (SELECT COUNT(*) FROM review_votes AS v WHERE v.review_id = r.id AND v.vote_type = 1),
			dislike_count = (SELECT COUNT(*) FROM review_votes AS v WHERE v.review_id = r.id AND v.vote_type = -1)
	`).Error; err != nil {
		return err
	}
	if err := db.Exec(`
		UPDATE courses AS c SET
			rating_count = (SELECT COUNT(*) FROM reviews AS r WHERE r.course_id = c.id),
			rating_avg = (SELECT COALESCE(AVG(rating), 0) FROM reviews AS r WHERE r.course_id = c.id)
	`).Error; err != nil {
		return err
	}
	if err := repository.RefreshTeacherSearchVectors(db); err != nil {
		return err
	}
	if err := repository.RefreshCourseSearchVectors(db); err != nil {
		return err
	}
	if err := repository.RefreshReviewSearchVectors(db); err != nil {
		return err
	}
	return m.checkpoint.MarkDone(stage)
}

func (m *Migrator) resetSequences() error {
	const stage = "reset_sequences"
	if m.checkpoint.StageDone(stage) {
		log.Println("Resetting sequences... already done, skipping")
		return nil
	}
	db := m.target
	log.Println("Resetting sequences...")
	for _, table := range []string{"teachers", "courses", "users", "user_point_records", "reviews", "review_revisions"} {
		if err := resetSequence(db, table, "id"); err != nil {
			return err
		}
	}
	return m.checkpoint.MarkDone(stage)
}

func resetSequence(db *gorm.DB, table, column string) error {
	return db.Exec(`
		SELECT setval(
			pg_get_serial_sequence(?, ?),
			COALESCE((SELECT MAX(id) FROM `+table+`), 1),
			COALESCE((SELECT MAX(id) FROM `+table+`), 0) > 0
		)
	`, table, column).Error
}

func userRole(row legacyUser) string {
	if row.IsStaff || row.IsSuperuser {
		return auth.RoleAdmin
	}
	return auth.RoleUser
}

func migratePasswordHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if hash == "" || strings.HasPrefix(hash, "!") {
		return ""
	}
	return hash
}

func legacyDepartmentName(department *legacyDepartment) string {
	if department == nil {
		return ""
	}
	return department.Name
}

func legacySemesterName(semester *legacySemester) string {
	if semester == nil {
		return ""
	}
	return semester.Name
}

func legacyCategoryNames(categories []legacyCategory) []string {
	if len(categories) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(categories))
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		name := strings.TrimSpace(category.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func legacyTeacherIDs(teachers []legacyTeacher) []int64 {
	if len(teachers) == 0 {
		return nil
	}
	seen := make(map[int]bool, len(teachers))
	ids := make([]int64, 0, len(teachers))
	for _, teacher := range teachers {
		if teacher.ID == 0 || seen[teacher.ID] {
			continue
		}
		seen[teacher.ID] = true
		ids = append(ids, int64(teacher.ID))
	}
	return ids
}

func generatePinyin(name string) (string, string) {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	py := pinyin.LazyPinyin(name, a)
	full := strings.Join(py, " ")

	a.Style = pinyin.FirstLetter
	abbrPy := pinyin.LazyPinyin(name, a)
	abbr := strings.Join(abbrPy, "")
	return full, abbr
}

func userLastSeen(row legacyUser, fallback time.Time) time.Time {
	if row.LastSeenAt.Valid {
		return row.LastSeenAt.Time
	}
	if row.LastLogin.Valid {
		return row.LastLogin.Time
	}
	return fallback
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func intOrZero(value sql.NullInt64) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int64)
}

func floatOrZero(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}
