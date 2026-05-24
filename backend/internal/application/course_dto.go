package application

import (
	"time"

	"jcourse/internal/domain/course"
)

// Read model: rating summary with distribution
type RatingInfoDTO struct {
	Count        int     `json:"count"`
	Avg          float64 `json:"avg"`
	Distribution [5]int  `json:"distribution"`
}

func newRatingInfoDTO(info course.RatingInfo) RatingInfoDTO {
	return RatingInfoDTO{
		Count:        info.Count,
		Avg:          info.Avg,
		Distribution: info.Distribution,
	}
}

// Read model: course list/search result
type CourseListItemDTO struct {
	ID          int           `json:"id"`
	Code        string        `json:"code"`
	Name        string        `json:"name"`
	Credit      float32       `json:"credit"`
	Department  string        `json:"department"`
	Language    string        `json:"language"`
	TargetYears []string      `json:"target_years"`
	Categories  []string      `json:"categories"`
	MainTeacher TeacherDTO    `json:"main_teacher"`
	Rating      RatingInfoDTO `json:"rating"`
}

func newCourseListItemDTO(c *course.CourseView) CourseListItemDTO {
	item := CourseListItemDTO{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Credit:      c.Credit,
		Department:  c.Department,
		Language:    c.Language,
		TargetYears: c.TargetYears,
		Categories:  c.Categories,
		Rating:      newRatingInfoDTO(c.Rating),
	}
	if c.MainTeacher != nil {
		item.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return item
}

// Read model: offered course
type OfferedCourseDTO struct {
	Semester    string   `json:"semester"`
	Language    string   `json:"language"`
	TargetYears []string `json:"target_years"`
	Categories  []string `json:"categories"`
}

// Read model: course detail with offered courses, related courses, and rating distribution
type CourseDetailDTO struct {
	ID                 int                   `json:"id"`
	Code               string                `json:"code"`
	Name               string                `json:"name"`
	Credit             float32               `json:"credit"`
	Department         string                `json:"department"`
	LastSemester       string                `json:"last_semester"`
	Language           string                `json:"language"`
	TargetYears        []string              `json:"target_years"`
	Categories         []string              `json:"categories"`
	MainTeacher        TeacherDTO            `json:"main_teacher"`
	TeacherGroup       []TeacherDTO          `json:"teacher_group,omitempty"`
	OfferedCourses     []OfferedCourseDTO    `json:"offered_courses"`
	Rating             RatingInfoDTO         `json:"rating"`
	SameCodeCourses    []CourseListItemDTO   `json:"same_code_courses"`
	SameTeacherCourses []CourseListItemDTO   `json:"same_teacher_courses"`
	NotificationLevel  int                   `json:"notification_level"`
	MyEnrollments      []CourseEnrollmentDTO `json:"my_enrollments,omitempty"`
	MyReview           *ReviewDTO            `json:"my_review,omitempty"`
}

type CourseEnrollmentDTO struct {
	ID        int               `json:"id"`
	Course    CourseListItemDTO `json:"course"`
	Semester  string            `json:"semester"`
	CreatedAt time.Time         `json:"created_at"`
}

func newCourseEnrollmentDTO(e *course.CourseEnrollmentView) CourseEnrollmentDTO {
	return CourseEnrollmentDTO{
		ID:        e.ID,
		Course:    newCourseListItemDTO(&e.Course),
		Semester:  e.Semester,
		CreatedAt: e.CreatedAt,
	}
}

type HotCourseItemDTO struct {
	Course CourseListItemDTO `json:"course"`
	Score  int64             `json:"score"`
}

type HotCourseListDTO struct {
	Period string             `json:"period"`
	Items  []HotCourseItemDTO `json:"items"`
}

type PaginatedResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
