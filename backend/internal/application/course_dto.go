package application

import "jcourse/internal/domain/course"

type CourseListItem struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Credit      float32    `json:"credit"`
	MainTeacher TeacherDTO `json:"main_teacher"`
	RatingCount int        `json:"rating_cnt"`
	RatingAvg   float64    `json:"rating_avg"`
}

func newCourseListItem(c *course.CourseForQuery) CourseListItem {
	item := CourseListItem{
		Code:        c.Code,
		Name:        c.Name,
		Credit:      c.Credit,
		RatingCount: c.ReviewCount,
		RatingAvg:   c.AvgRating,
	}
	if c.MainTeacher != nil {
		item.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return item
}

type OfferedCourseDTO struct {
	Semester     string       `json:"semester"`
	Language     string       `json:"language"`
	Grade        string       `json:"grade"`
	TeacherGroup []TeacherDTO `json:"teacher_group"`
}

type CourseDetailDTO struct {
	ID                 int                `json:"id"`
	Code               string             `json:"code"`
	Name               string             `json:"name"`
	Credit             float32            `json:"credit"`
	Department         string             `json:"department"`
	MainTeacher        TeacherDTO         `json:"main_teacher"`
	OfferedCourses     []OfferedCourseDTO `json:"offered_courses"`
	ReviewCount        int                `json:"review_count"`
	AvgRating          float64            `json:"avg_rating"`
	RatingDistribution [5]int             `json:"rating_distribution"`
	OtherTeachers      []CourseListItem   `json:"other_teachers"`
	OtherCourses       []CourseListItem   `json:"other_courses"`
}

type PaginatedResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
