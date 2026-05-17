package web

import (
	"github.com/gin-gonic/gin"

	"jcourse/internal/app"
	"jcourse/internal/interface/web/controller"
)

func NewRouter(container *app.ServiceContainer) *gin.Engine {
	g := gin.Default()

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
