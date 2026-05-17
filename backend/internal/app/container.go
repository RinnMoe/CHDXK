package app

import (
	"jcourse/config"
	"jcourse/internal/application"
	"jcourse/internal/infrastructure/persistence"
	"jcourse/internal/infrastructure/repository"
)

type ServiceContainer struct {
	ReviewQuery   *application.ReviewQueryService
	ReviewCommand *application.ReviewCommandService
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)

	reviewRepo := repository.NewReviewRepository(db)
	courseRepo := repository.NewCourseRepository(db)

	reviewQuery := application.NewReviewQueryService(reviewRepo)
	reviewCommand := application.NewReviewCommandService(courseRepo, reviewRepo, nil)

	return &ServiceContainer{
		ReviewQuery:   reviewQuery,
		ReviewCommand: reviewCommand,
	}
}
