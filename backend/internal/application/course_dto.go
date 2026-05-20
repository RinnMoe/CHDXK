package application

import "jcourse/internal/domain/course"

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

type CourseListItem struct {
	ID          int           `json:"id"`
	Code        string        `json:"code"`
	Name        string        `json:"name"`
	Credit      float32       `json:"credit"`
	Language    string        `json:"language"`
	Grades      []string      `json:"grades"`
	Categories  []string      `json:"categories"`
	MainTeacher TeacherDTO    `json:"main_teacher"`
	Rating      RatingInfoDTO `json:"rating"`
}

func newCourseListItem(c *course.CourseForQuery) CourseListItem {
	item := CourseListItem{
		ID:         c.ID,
		Code:       c.Code,
		Name:       c.Name,
		Credit:     c.Credit,
		Language:   c.Language,
		Grades:     c.Grades,
		Categories: c.Categories,
		Rating:     newRatingInfoDTO(c.Rating),
	}
	if c.MainTeacher != nil {
		item.MainTeacher = newTeacherDTO(c.MainTeacher)
	}
	return item
}

type OfferedCourseDTO struct {
	Semester     string       `json:"semester"`
	Language     string       `json:"language"`
	Grades       []string     `json:"grades"`
	Categories   []string     `json:"categories"`
	TeacherGroup []TeacherDTO `json:"teacher_group"`
}

type CourseDetailDTO struct {
	ID             int                `json:"id"`
	Code           string             `json:"code"`
	Name           string             `json:"name"`
	Credit         float32            `json:"credit"`
	Department     string             `json:"department"`
	Language       string             `json:"language"`
	Grades         []string           `json:"grades"`
	Categories     []string           `json:"categories"`
	MainTeacher    TeacherDTO         `json:"main_teacher"`
	OfferedCourses []OfferedCourseDTO `json:"offered_courses"`
	Rating         RatingInfoDTO      `json:"rating"`
	OtherTeachers  []CourseListItem   `json:"other_teachers"`
	OtherCourses   []CourseListItem   `json:"other_courses"`
}

type PaginatedResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}
