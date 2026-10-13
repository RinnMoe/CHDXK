package application

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
)

type CourseDiscoveryDTO struct {
	Semester string              `json:"semester"`
	Items    []CourseListItemDTO `json:"items"`
}

type CourseReviewPromptDTO struct {
	Course   CourseListItemDTO `json:"course"`
	Semester string            `json:"semester"`
}

type CourseReviewPromptsDTO struct {
	HasEnrollments bool                    `json:"has_enrollments"`
	Items          []CourseReviewPromptDTO `json:"items"`
}

var academicSemesterPattern = regexp.MustCompile(`^(\d{4})-(\d{4})-([12])$`)

func latestCatalogSemester(filters *course.CourseFilters) string {
	latest := ""
	if filters == nil {
		return latest
	}
	for _, semester := range filters.Semesters {
		parts := academicSemesterPattern.FindStringSubmatch(semester.Name)
		if parts == nil {
			continue
		}
		start, _ := strconv.Atoi(parts[1])
		end, _ := strconv.Atoi(parts[2])
		if end == start+1 && semester.Name > latest {
			latest = semester.Name
		}
	}
	return latest
}

func (s *CourseQueryService) DiscoverCourses(ctx context.Context) (*CourseDiscoveryDTO, error) {
	filters, err := s.courseRepo.GetFilters(ctx)
	if err != nil {
		return nil, err
	}
	semester := latestCatalogSemester(filters)
	date := time.Now().In(s.hotCourseLocation).Format("2006-01-02")
	rows, err := s.courseRepo.FindDiscovery(ctx, semester, date, 24)
	if err != nil {
		return nil, err
	}
	result := &CourseDiscoveryDTO{Semester: semester, Items: []CourseListItemDTO{}}
	for _, row := range rows {
		result.Items = append(result.Items, newCourseListItemDTO(&row))
	}
	return result, nil
}

func (s *CourseQueryService) ReviewPrompts(ctx context.Context, userID int) (*CourseReviewPromptsDTO, error) {
	rows, err := s.enrollmentQuery.FindUserEnrollments(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Filter at the use-case boundary as well as the repository boundary.
	enrollments := make([]course.CourseEnrollmentView, 0, len(rows))
	for _, row := range rows {
		if row.UserID == userID {
			enrollments = append(enrollments, row)
		}
	}
	result := &CourseReviewPromptsDTO{HasEnrollments: len(enrollments) > 0, Items: []CourseReviewPromptDTO{}}
	if !result.HasEnrollments {
		return result, nil
	}
	sort.Slice(enrollments, func(i, j int) bool {
		if enrollments[i].Semester != enrollments[j].Semester {
			return enrollments[i].Semester > enrollments[j].Semester
		}
		return enrollments[i].Course.ID < enrollments[j].Course.ID
	})
	ids := make([]int, 0, len(enrollments))
	seen := map[int]bool{}
	for _, row := range enrollments {
		if !seen[row.Course.ID] {
			ids = append(ids, row.Course.ID)
			seen[row.Course.ID] = true
		}
	}
	reviews, _, err := s.reviewQuery.FindBy(ctx, review.ReviewFilter{UserID: userID, CourseIDs: ids})
	if err != nil {
		return nil, err
	}
	reviewed := map[int]bool{}
	for _, row := range reviews {
		if row.UserID == userID {
			reviewed[row.CourseID] = true
		}
	}
	seen = map[int]bool{}
	for _, row := range enrollments {
		if reviewed[row.Course.ID] || seen[row.Course.ID] {
			continue
		}
		seen[row.Course.ID] = true
		result.Items = append(result.Items, CourseReviewPromptDTO{Course: newCourseListItemDTO(&row.Course), Semester: row.Semester})
		if len(result.Items) == 3 {
			break
		}
	}
	return result, nil
}
