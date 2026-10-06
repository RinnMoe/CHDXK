package course

import (
	"context"
	"testing"
)

func TestManualEnrollmentValidation(t *testing.T) {
	calls := 0
	repo := &MockCourseEnrollmentRepository{OnSave: func(_ context.Context, id, owner, section int) error {
		calls++
		if id != 0 || owner != 1 || section != 2 {
			t.Fatal("wrong owner or section")
		}
		return nil
	}}
	service := NewEnrollmentService(repo)
	for _, args := range [][3]int{{0, 0, 2}, {1, -1, 2}, {1, 0, 0}} {
		if err := service.Save(context.Background(), args[0], args[1], args[2]); err == nil {
			t.Fatal("invalid selection accepted")
		}
	}
	if calls != 0 {
		t.Fatal("invalid request reached repository")
	}
	if err := service.Save(context.Background(), 1, 0, 2); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("valid request not saved")
	}
}
