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

	apiGroup := g.Group("/api")
	courseGroup := apiGroup.Group("/course")
	{
		courseGroup.GET("/:courseID/review", reviewController.GetCourseReviews)
	}
	reviewGroup := apiGroup.Group("/review")
	{
		reviewGroup.GET("/latest", reviewController.GetLatestReviews)
		reviewGroup.GET("/:reviewID", reviewController.GetReviewDetail)
		reviewGroup.POST("/", reviewController.CreateReview)
		reviewGroup.PUT("/:reviewID", reviewController.UpdateReview)
		reviewGroup.DELETE("/:reviewID", reviewController.DeleteReview)
	}
	userGroup := apiGroup.Group("/user")
	{
		userGroup.GET("/:userID/reviews", reviewController.GetMyReviews)
	}

	return g
}
