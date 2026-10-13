package repository_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"jcourse/internal/domain/course"
	"jcourse/internal/infrastructure/repository"
)

func TestCourseDiscoveryDistributionAndRotation(t *testing.T) {
	db := newTestDB(t)
	teacher := seedTeacher(t, db)
	for department := 0; department < 4; department++ {
		for i := 0; i < 9; i++ {
			c := seedCourseRaw(t, db, fmt.Sprintf("D%dC%d", department, i), "Course", 2, fmt.Sprintf("Department %d", department), teacher.ID, "zh", nil, nil)
			seedOfferedCourse(t, db, c.ID, "2026-2027-1")
			seedOfferedCourse(t, db, c.ID, "2025-2026-2")
		}
	}
	old := seedCourseRaw(t, db, "OLD", "Old course", 3, "Old department", teacher.ID, "zh", nil, nil)
	seedOfferedCourse(t, db, old.ID, "2025-2026-2")
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()
	get := func(date string, limit int) []course.CourseView {
		rows, err := repo.FindDiscovery(ctx, "2026-2027-1", date, limit)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	first := get("2026-10-10", 24)
	if len(first) != 24 || !reflect.DeepEqual(first, get("2026-10-10", 24)) {
		t.Fatal("limit or daily stability failed")
	}
	if reflect.DeepEqual(first, get("2026-10-11", 24)) {
		t.Fatal("daily rotation failed")
	}
	seen := map[int]bool{}
	for start := 0; start < len(first); start += 4 {
		departments := map[string]bool{}
		for _, row := range first[start : start+4] {
			if seen[row.ID] || row.ID == old.ID {
				t.Fatal("duplicate or wrong semester course")
			}
			seen[row.ID] = true
			departments[row.Department] = true
		}
		if len(departments) != 4 {
			t.Fatal("departments not interleaved")
		}
	}
	if len(get("2026-10-10", 3)) != 3 || len(get("2026-10-10", 0)) != 0 {
		t.Fatal("small limits failed")
	}
	missing, err := repo.FindDiscovery(ctx, "missing", "2026-10-10", 24)
	if err != nil || len(missing) != 0 {
		t.Fatalf("empty semester = %v %v", missing, err)
	}
	// Rating changes must not influence discovery selection.
	if err := db.Model(&repository.CourseEntity{}).Where("id = ?", first[0].ID).Updates(map[string]any{"rating_score": 5, "rating_count": 200}).Error; err != nil {
		t.Fatal(err)
	}
	updated := get("2026-10-10", 24)
	for i := range first {
		if first[i].ID != updated[i].ID {
			t.Fatal("ratings altered discovery")
		}
	}
}

func TestCourseDiscoverySmallAndEmptyCatalog(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewCourseRepository(db)
	ctx := context.Background()
	empty, err := repo.FindDiscovery(ctx, "", "2026-10-10", 24)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty catalog = %v %v", empty, err)
	}
	teacher := seedTeacher(t, db)
	seedCourse(t, db, teacher.ID)
	rows, err := repo.FindDiscovery(ctx, "", "2026-10-10", 24)
	if err != nil || len(rows) != 1 {
		t.Fatalf("small fallback catalog = %v %v", rows, err)
	}
}
