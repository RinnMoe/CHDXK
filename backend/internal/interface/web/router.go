package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"jcourse/config"
	"jcourse/internal/app"
	"jcourse/internal/interface/web/controller"
	"jcourse/internal/interface/web/middleware"
)

func NewRouter(container *app.ServiceContainer, conf config.AppConfig) *gin.Engine {
	g := gin.Default()

	store, err := middleware.NewSessionStore(conf.Redis, conf.Session)
	if err != nil {
		panic(err)
	}
	g.Use(sessions.Sessions("jcourse_session", store))

	reviewController := controller.NewReviewController(container.ReviewQuery, container.ReviewCommand)
	courseController := controller.NewCourseController(container.CourseQuery, container.CourseCommand)
	teacherController := controller.NewTeacherController(container.TeacherQuery, container.CourseQuery)

	apiGroup := g.Group("/api")
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/", courseController.ListCourses)
		courseGroup.GET("/followed", courseController.ListFollowedCourses)
		courseGroup.GET("/ignored", courseController.ListIgnoredCourses)
		courseGroup.GET("/:courseID", courseController.GetCourseDetail)
		courseGroup.GET("/:courseID/review", reviewController.GetCourseReviews)
		courseGroup.POST("/:courseID/notification", courseController.SetNotificationLevel)
	}
	teacherGroup := apiGroup.Group("/teacher")
	{
		teacherGroup.GET("/", teacherController.ListTeachers)
		teacherGroup.GET("/:teacherID/courses", teacherController.ListTeacherCourses)
	}
	reviewGroup := apiGroup.Group("/review")
	{
		reviewGroup.GET("/latest", reviewController.GetLatestReviews)
		reviewGroup.GET("/followed", reviewController.GetFollowedReviews)
		reviewGroup.GET("/:reviewID", reviewController.GetReviewDetail)
		reviewGroup.POST("/", reviewController.CreateReview)
		reviewGroup.POST("/:reviewID/vote", reviewController.VoteReview)
		reviewGroup.PUT("/:reviewID", reviewController.UpdateReview)
		reviewGroup.DELETE("/:reviewID", reviewController.DeleteReview)
	}
	userGroup := apiGroup.Group("/user")
	{
		userGroup.GET("/:userID/reviews", reviewController.GetMyReviews)
	}

	return g
}
