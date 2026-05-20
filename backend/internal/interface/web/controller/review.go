package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/review"
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

func (r *ReviewController) VoteReview(c *gin.Context) {
	reviewID, err := strconv.Atoi(c.Param("reviewID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var cmd application.VoteCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if cmd.VoteType != review.VoteLike && cmd.VoteType != review.VoteDislike && cmd.VoteType != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vote_type must be 1, -1, or 0"})
		return
	}

	u := auth.GetUserFromCtx(c.Request.Context())
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := r.command.VoteReview(c.Request.Context(), u.ID, reviewID, cmd.VoteType); err != nil {
		if errors.Is(err, review.ErrDailyVoteLimitReached) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
