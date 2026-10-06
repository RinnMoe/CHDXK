package course

import "context"

type EnrollmentSourceCourse struct {
	LessonNo       string `json:"lesson_no"`
	SourceLessonID string `json:"source_lesson_id"`
	Name           string `json:"name"`
}
type EnrollmentImportResult struct {
	Matched   int                      `json:"matched"`
	Added     int                      `json:"added"`
	Unmatched []EnrollmentSourceCourse `json:"unmatched"`
}
type EnrollmentImporter interface {
	ImportSemester(context.Context, int, string, []EnrollmentSourceCourse) (*EnrollmentImportResult, error)
}
