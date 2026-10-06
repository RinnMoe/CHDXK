package repository_test

import (
	"context"
	"jcourse/internal/infrastructure/repository"
	"testing"
)

func TestManualEnrollmentUsesExistingSectionAndOwnsRecord(t *testing.T) {
	db := newTestDB(t)
	teacher := seedTeacher(t, db)
	c := seedCourse(t, db, teacher.ID)
	u := seedUser(t, db)
	other := seedUserRaw(t, db, "manual-other", "manual-other@example.com")
	first := seedOfferedCourseRaw(t, db, c.ID, "2026-2027-1", "中文", nil, nil)
	second := seedOfferedCourseRaw(t, db, c.ID, "2025-2026-2", "中文", nil, nil)
	a := repository.CourseSectionEntity{OfferedCourseID: first.ID, LessonNo: "TEST.01"}
	b := repository.CourseSectionEntity{OfferedCourseID: second.ID, LessonNo: "TEST.02"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewCourseEnrollmentRepository(db)
	ctx := context.Background()
	if err := repo.Save(ctx, 0, u.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	rows, err := repo.FindUserEnrollments(ctx, u.ID)
	if err != nil || len(rows) != 1 || rows[0].Course.ID != c.ID || rows[0].Semester != first.Semester {
		t.Fatalf("wrong section mapping: %+v %v", rows, err)
	}
	id := rows[0].ID
	if err := repo.Save(ctx, 0, u.ID, a.ID); err == nil {
		t.Fatal("duplicate accepted")
	}
	if err := repo.Save(ctx, id, other.ID, b.ID); err == nil {
		t.Fatal("another user edited record")
	}
	if err := repo.Save(ctx, 0, u.ID, -1); err == nil {
		t.Fatal("nonexistent section accepted")
	}
	if err := repo.Save(ctx, id, u.ID, b.ID); err != nil {
		t.Fatal(err)
	}
	rows, err = repo.FindUserEnrollments(ctx, u.ID)
	if err != nil || len(rows) != 1 || rows[0].ID != id || rows[0].Semester != second.Semester {
		t.Fatalf("edit failed: %+v %v", rows, err)
	}
	if err := repo.Save(ctx, 0, u.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, id, u.ID, a.ID); err == nil {
		t.Fatal("edit to duplicate accepted")
	}
	rows, err = repo.FindUserEnrollments(ctx, u.ID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("failed edit damaged records: %+v %v", rows, err)
	}
}
