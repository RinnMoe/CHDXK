package repository

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"jcourse/internal/domain/course"
	"jcourse/internal/domain/review"
	"jcourse/internal/domain/teacher"
)

const findByCountCacheTTL = time.Minute

func reviewFindByCountCacheKey(filter review.ReviewFilter) string {
	return cacheKey("review", "find_by", "count", countCacheDigest(
		filter.ReviewID,
		filter.CourseID,
		joinInts(filter.CourseIDs),
		joinInts(filter.ExcludeCourseIDs),
		filter.UserID,
		filter.Q,
		filter.Semester,
		filter.Rating,
		filter.CreatedAfter.UTC().Format(time.RFC3339Nano),
		filter.WithCourse,
	))
}

func courseFindByCountCacheKey(filter course.CourseFilter) string {
	hasReview := ""
	if filter.HasReview != nil {
		hasReview = strconv.FormatBool(*filter.HasReview)
	}

	credit := ""
	if filter.Credit != nil {
		credit = strconv.FormatFloat(float64(*filter.Credit), 'f', -1, 32)
	}

	return cacheKey("course", "find_by", "count", countCacheDigest(
		joinInts(filter.CourseIDs),
		filter.TeacherID,
		filter.ExcludeID,
		filter.Q,
		filter.Code,
		filter.Name,
		filter.MainTeacherName,
		filter.Department,
		strings.Join(filter.Categories, ","),
		filter.Language,
		strings.Join(filter.TargetYears, ","),
		credit,
		hasReview,
	))
}

func teacherFindByCountCacheKey(filter teacher.TeacherFilter) string {
	return cacheKey("teacher", "find_by", "count", countCacheDigest(
		joinInts(filter.TeacherIDs),
		filter.Department,
		filter.Title,
		filter.Q,
		filter.Code,
		filter.Name,
	))
}

func countCacheDigest(parts ...any) string {
	h := fnv.New64a()
	for _, part := range parts {
		fmt.Fprint(h, part)
		h.Write([]byte{0})
	}
	return strconv.FormatUint(h.Sum64(), 16)
}

func joinInts(values []int) string {
	if len(values) == 0 {
		return ""
	}

	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprint(value)
	}
	return strings.Join(parts, ",")
}
