package main

import (
	"log"

	"jcourse/config"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"

	"github.com/spf13/pflag"
)

func main() {
	configPath := pflag.String("config", "config/config.yaml", "path to config file")
	pflag.Parse()

	conf, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db := persistence.NewPostgres(conf.Postgres)

	err = db.AutoMigrate(
		&repository.DepartmentEntity{},
		&repository.SemesterEntity{},
		&repository.TeacherEntity{},
		&repository.CourseEntity{},
		&repository.OfferedCourseEntity{},
		&repository.CourseTeacherGroupEntity{},
		&repository.UserEntity{},
		&repository.ReviewEntity{},
		&repository.ReviewRevisionEntity{},
	)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("migration complete")
}
