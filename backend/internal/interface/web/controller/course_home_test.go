package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"jcourse/internal/application"
	"jcourse/internal/domain/auth"
	"jcourse/internal/domain/course"
)

type homeEnrollmentQuery struct {
	course.CourseEnrollmentQuery
	requestedUserID int
}

func (q *homeEnrollmentQuery) FindUserEnrollments(ctx context.Context, userID int) ([]course.CourseEnrollmentView, error) {
	q.requestedUserID = userID
	return []course.CourseEnrollmentView{}, nil
}

func TestReviewPromptsUsesSessionIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, signedIn := range []bool{false, true} {
		q := &homeEnrollmentQuery{}
		service := application.NewCourseQueryService(nil, nil, nil, nil, q, nil)
		ctrl := NewCourseController(service, nil)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/course/review-prompts?user_id=999", nil)
		if signedIn {
			c.Request = c.Request.WithContext(auth.WithUser(c.Request.Context(), &auth.User{ID: 7}))
		}
		ctrl.ReviewPrompts(c)
		if !signedIn {
			if recorder.Code != http.StatusUnauthorized || q.requestedUserID != 0 {
				t.Fatal("unauthenticated access was not rejected")
			}
			continue
		}
		if recorder.Code != http.StatusOK || q.requestedUserID != 7 {
			t.Fatalf("status %d, requested user %d", recorder.Code, q.requestedUserID)
		}
		var dto application.CourseReviewPromptsDTO
		if err := json.Unmarshal(recorder.Body.Bytes(), &dto); err != nil || dto.HasEnrollments || dto.Items == nil {
			t.Fatalf("invalid empty result: %s", recorder.Body.String())
		}
	}
}
