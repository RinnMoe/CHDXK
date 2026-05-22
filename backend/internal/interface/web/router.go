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
	pointController := controller.NewPointController(container.PointQuery, container.PointCommand)
	authController := controller.NewAuthController(container.AuthCommand)
	siteStatsController := controller.NewSiteStatsController(container.SiteStatsQuery)
	announcementController := controller.NewAnnouncementController(container.AnnouncementQuery)

	apiGroup := g.Group("/api")
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register/code", authController.SendRegisterCode)
		authGroup.POST("/register", authController.Register)
		authGroup.POST("/login", authController.Login)
		authGroup.POST("/logout", authController.Logout)
		authGroup.GET("/me", authController.Me)
		authGroup.POST("/password-reset/code", authController.SendResetCode)
		authGroup.POST("/password-reset", authController.ResetPassword)
	}
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/filters", courseController.GetCourseFilters)
		courseGroup.GET("/", courseController.ListCourses)
		courseGroup.GET("/followed", courseController.ListFollowedCourses)
		courseGroup.GET("/ignored", courseController.ListIgnoredCourses)
		courseGroup.GET("/:courseID", courseController.GetCourse)
		courseGroup.GET("/:courseID/review/filters", reviewController.GetCourseReviewFilters)
		courseGroup.GET("/:courseID/review", reviewController.ListCourseReviews)
		courseGroup.POST("/:courseID/notification", courseController.SetNotificationLevel)
	}
	teacherGroup := apiGroup.Group("/teacher")
	{
		teacherGroup.GET("/filters", teacherController.GetTeacherFilters)
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
		userGroup.GET("/:userID/points", pointController.GetUserPoints)
		userGroup.GET("/:userID/reviews", reviewController.ListUserReviews)
	}
	pointGroup := apiGroup.Group("/point")
	{
		pointGroup.POST("/transfers/preview", pointController.PreviewTransfer)
		pointGroup.POST("/transfers", pointController.CreateTransfer)
	}
	siteStatsGroup := apiGroup.Group("/site-stats", middleware.Admin())
	{
		siteStatsGroup.GET("/daily/yesterday", siteStatsController.GetYesterday)
		siteStatsGroup.GET("/daily", siteStatsController.ListDaily)
	}
	announcementGroup := apiGroup.Group("/announcement")
	{
		announcementGroup.GET("/", announcementController.ListAnnouncements)
	}

	extGroup := apiGroup.Group("/ext", middleware.APIKeyAuth(container.ApiKeySvc))
	{
		extGroup.GET("/points", pointController.GetPointsByEmail)
	}

	return g
}
