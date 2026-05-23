package repository_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"jcourse/internal/infrastructure/repository"
)

const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "postgres"
	testDBPassword = "postgres"
	testDBName     = "jcourse_test"
)

func testDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName,
	)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	rootDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword,
	)

	rootDB, err := gorm.Open(postgres.Open(rootDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	sqlDB, _ := rootDB.DB()
	defer sqlDB.Close()

	rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err := rootDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName)).Error; err != nil {
		t.Fatalf("create test database: %v", err)
	}

	db, err := gorm.Open(postgres.Open(testDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_jieba;").Error; err != nil {
		t.Fatalf("create pg_jieba extension: %v", err)
	}

	migrateTestDB(t, db)

	t.Cleanup(func() {
		sqlDB2, _ := db.DB()
		sqlDB2.Close()

		rootDB2, err := gorm.Open(postgres.Open(rootDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Printf("cleanup: connect root db: %v", err)
			return
		}
		defer func() {
			s, _ := rootDB2.DB()
			s.Close()
		}()
		rootDB2.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", testDBName))
	})

	return db
}

func migrateTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	err := db.AutoMigrate(
		&repository.DepartmentEntity{},
		&repository.SemesterEntity{},
		&repository.TeacherEntity{},
		&repository.CourseEntity{},
		&repository.OfferedCourseEntity{},
		&repository.UserEntity{},
		&repository.UserPointRecordEntity{},
		&repository.PointTransferEntity{},
		&repository.ReviewEntity{},
		&repository.ReviewRevisionEntity{},
		&repository.ReviewVoteEntity{},
		&repository.CourseNotificationEntity{},
		&repository.CourseHotScoreEntity{},
		&repository.SiteDailyStatEntity{},
		&repository.AnnouncementEntity{},
	)
	if err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func cleanTables(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()
	for _, tbl := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tbl)).Error; err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
}

func seedTeacher(t *testing.T, db *gorm.DB) repository.TeacherEntity {
	t.Helper()
	e := repository.TeacherEntity{
		Code:       "T001",
		Name:       "张三",
		Department: "计算机学院",
		Title:      "教授",
		Pinyin:     "zhangsan",
		PinyinAbbr: "zs",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}
	return e
}

func seedCourse(t *testing.T, db *gorm.DB, teacherID int) repository.CourseEntity {
	t.Helper()
	e := repository.CourseEntity{
		Code:          "CS101",
		Name:          "数据结构",
		Credit:        3.0,
		Department:    "计算机学院",
		MainTeacherID: teacherID,
		Categories:    []string{"核心课", "必修课"},
		Language:      "zh",
		TargetYears:   []string{"2021", "2022"},
		RatingCount:   0,
		RatingAvg:     0,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed course: %v", err)
	}
	return e
}

func seedOfferedCourse(t *testing.T, db *gorm.DB, courseID int, semester string) repository.OfferedCourseEntity {
	t.Helper()
	e := repository.OfferedCourseEntity{
		CourseID:    courseID,
		Semester:    semester,
		Language:    "zh",
		TargetYears: []string{"2022"},
		Categories:  []string{"核心课"},
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed offered course: %v", err)
	}
	return e
}

func seedCourseRaw(t *testing.T, db *gorm.DB, code, name string, credit float32, department string, teacherID int, language string, categories, targetYears []string) repository.CourseEntity {
	t.Helper()
	e := repository.CourseEntity{
		Code:          code,
		Name:          name,
		Credit:        credit,
		Department:    department,
		MainTeacherID: teacherID,
		Categories:    categories,
		Language:      language,
		TargetYears:   targetYears,
		RatingCount:   0,
		RatingAvg:     0,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed course: %v", err)
	}
	return e
}

func seedOfferedCourseRaw(t *testing.T, db *gorm.DB, courseID int, semester, language string, categories, targetYears []string) repository.OfferedCourseEntity {
	t.Helper()
	e := repository.OfferedCourseEntity{
		CourseID:    courseID,
		Semester:    semester,
		Language:    language,
		TargetYears: targetYears,
		Categories:  categories,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed offered course: %v", err)
	}
	return e
}

func seedUser(t *testing.T, db *gorm.DB) repository.UserEntity {
	t.Helper()
	e := repository.UserEntity{
		Username:   "testuser",
		Email:      "testuser@example.com",
		Role:       "user",
		Password:   "hashed_password",
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return e
}

func seedReview(t *testing.T, db *gorm.DB, courseID, userID int) repository.ReviewEntity {
	t.Helper()
	e := repository.ReviewEntity{
		CourseID:  courseID,
		Semester:  "2024-2025-1",
		UserID:    userID,
		Rating:    5,
		Content:   "很好的课程，老师讲解清晰。",
		Score:     "A",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatalf("seed review: %v", err)
	}
	return e
}

var _ = context.Background
