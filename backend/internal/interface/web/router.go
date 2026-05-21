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
	g.Use(middleware.OptionalAuth(container.AuthService))

	reviewController := controller.NewReviewController(container.ReviewQuery, container.ReviewCommand)
	courseController := controller.NewCourseController(container.CourseQuery, container.CourseCommand)
	teacherController := controller.NewTeacherController(container.TeacherQuery, container.CourseQuery)
	authController := controller.NewAuthController(container.AuthCommand)

	apiGroup := g.Group("/api")
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register/code", authController.SendRegisterCode)
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", authController.Logout)
	}
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/", courseController.ListCourses)
		courseGroup.GET("/followed", courseController.ListFollowedCourses)
		courseGroup.GET("/ignored", courseController.ListIgnoredCourses)
		courseGroup.GET("/:courseID", courseController.GetCourse)
		courseGroup.GET("/:courseID/review", reviewController.ListCourseReviews)
		courseGroup.POST("/:courseID/notification", courseController.SetNotificationLevel)
	}
	teacherGroup := apiGroup.Group("/teacher")
	{
		teacherGroup.GET("/", teacherController.ListTeachers)
		teacherGroup.GET("/:teacherID/courses", teacherController.ListTeacherCourses)
	}
	reviewGroup := apiGroup.Group("/review")
	{
		reviewGroup.GET("/latest", reviewController.ListLatestReviews)
		reviewGroup.GET("/followed", reviewController.ListFollowedReviews)
		reviewGroup.GET("/:reviewID", reviewController.GetReview)
		reviewGroup.POST("/", reviewController.CreateReview)
		reviewGroup.POST("/:reviewID/vote", reviewController.VoteReview)
		reviewGroup.PUT("/:reviewID", reviewController.UpdateReview)
		reviewGroup.DELETE("/:reviewID", reviewController.DeleteReview)
	}
	userGroup := apiGroup.Group("/user")
	{
		userGroup.GET("/:userID/reviews", reviewController.ListUserReviews)
	}

	return g
}
