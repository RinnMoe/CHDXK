package jaccount

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

var (
	ErrDisabled         = errors.New("jaccount course sync disabled")
	ErrSemesterRequired = errors.New("semester required")
)

type LessonCourse struct {
	Code        string
	TeacherName string
}

type Client interface {
	AuthCodeURL(state string) (string, error)
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	Lessons(ctx context.Context, token *oauth2.Token, semester string) ([]LessonCourse, error)
}
