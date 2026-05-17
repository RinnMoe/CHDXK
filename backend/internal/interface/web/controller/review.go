package controller

import (
	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
)

type ReviewController struct {
	query   *application.ReviewQueryService
	command *application.ReviewCommandService
}

func NewReviewController(
	query *application.ReviewQueryService,
	command *application.ReviewCommandService,
) *ReviewController {
	return &ReviewController{query: query, command: command}
}

func (r *ReviewController) GetCourseReviews(c *gin.Context) {

}

func (r *ReviewController) CreateReview(c *gin.Context) {

}

func (r *ReviewController) UpdateReview(c *gin.Context) {

}

func (r *ReviewController) DeleteReview(c *gin.Context) {

}

func (r *ReviewController) GetMyReviews(c *gin.Context) {

}

func (r *ReviewController) GetLatestReviews(c *gin.Context) {

}

func (r *ReviewController) GetReviewDetail(c *gin.Context) {

}
