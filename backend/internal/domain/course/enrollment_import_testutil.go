//go:build test

package course

import "context"

type MockEnrollmentImporter struct {
	ImportSemesterFn func(context.Context, int, string, []EnrollmentSourceCourse) (*EnrollmentImportResult, error)
}

func (m *MockEnrollmentImporter) ImportSemester(ctx context.Context, user int, semester string, courses []EnrollmentSourceCourse) (*EnrollmentImportResult, error) {
	return m.ImportSemesterFn(ctx, user, semester, courses)
}
