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
	CourseQuery   *application.CourseQueryService
	TeacherQuery  *application.TeacherQueryService
}

func NewServiceContainer(conf config.AppConfig) *ServiceContainer {

	db := persistence.NewPostgres(conf.Postgres)

	reviewRepo := repository.NewReviewRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	teacherRepo := repository.NewTeacherRepository(db)

	reviewQuery := application.NewReviewQueryService(reviewRepo)
	reviewCommand := application.NewReviewCommandService(courseRepo, reviewRepo, nil)
	courseQuery := application.NewCourseQueryService(courseRepo)
	teacherQuery := application.NewTeacherQueryService(teacherRepo)

	return &ServiceContainer{
		ReviewQuery:   reviewQuery,
		ReviewCommand: reviewCommand,
		CourseQuery:   courseQuery,
		TeacherQuery:  teacherQuery,
	}
}
