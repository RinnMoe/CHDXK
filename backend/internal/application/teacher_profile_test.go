//go:build test

package application

import (
	"context"
	"errors"
	"jcourse/internal/domain/teacher"
	"testing"
)

func TestTeacherProfileSyncFailureNeverSavesPartialCatalog(t *testing.T) {
	saved := false
	repo := &teacher.MockProfileStore{OnTeachers: func(context.Context) ([]teacher.TeacherView, error) {
		return []teacher.TeacherView{{Name: "张三", Department: "理学院"}}, nil
	}, OnSave: func(context.Context, []teacher.TeacherView, []teacher.Profile) error { saved = true; return nil }}
	source := &teacher.MockProfileSource{OnList: func(context.Context, func(int, int)) ([]teacher.Profile, error) {
		return nil, errors.New("network failure")
	}}
	s := NewTeacherProfileService(repo, source)
	ctx, cancel := context.WithCancel(context.Background())
	s.run(ctx, cancel)
	if saved || s.Status().State != "failed" {
		t.Fatal(s.Status(), saved)
	}
}
func TestTeacherProfileSyncReadsOnlyMatchingColleges(t *testing.T) {
	saved := false
	repo := &teacher.MockProfileStore{OnTeachers: func(context.Context) ([]teacher.TeacherView, error) {
		return []teacher.TeacherView{{Name: "张三", Department: "理学院"}}, nil
	}, OnSave: func(_ context.Context, _ []teacher.TeacherView, p []teacher.Profile) error {
		saved = true
		if len(p) != 2 || p[0].Fields["研究领域"] != "数学" {
			t.Error("profile lost")
		}
		return nil
	}}
	source := &teacher.MockProfileSource{OnList: func(_ context.Context, progress func(int, int)) ([]teacher.Profile, error) {
		progress(1, 1)
		return []teacher.Profile{{Name: "张三", Department: "理学院"}, {Name: "张三", Department: "信息工程学院"}}, nil
	}, OnDetail: func(_ context.Context, p teacher.Profile) (teacher.Profile, error) {
		if p.Department != "理学院" {
			t.Error("different college detail requested")
		}
		p.Fields = map[string]string{"研究领域": "数学"}
		return p, nil
	}}
	s := NewTeacherProfileService(repo, source)
	ctx, cancel := context.WithCancel(context.Background())
	s.run(ctx, cancel)
	if !saved || s.Status().State != "done" || s.Status().Completed != 1 {
		t.Fatal(s.Status(), saved)
	}
}

func TestTeacherProfileEditValidation(t *testing.T) {
	saved := false
	repo := &teacher.MockProfileStore{OnEdit: func(_ context.Context, _ int, p teacher.Profile) error {
		saved = true
		if p.Fields["办公地点"] != "办公室" {
			t.Fatal(p)
		}
		return nil
	}}
	s := NewTeacherProfileService(repo, nil)
	for _, p := range []teacher.Profile{{Name: "", Department: "理学院"}, {Name: "张三", Department: "理学院", URL: "javascript:alert(1)"}, {Name: "张三", Department: "理学院", Fields: map[string]string{"密码": "x"}}} {
		if err := s.Edit(context.Background(), 1, p); err == nil {
			t.Fatal("invalid edit accepted")
		}
	}
	if saved {
		t.Fatal("invalid edit saved")
	}
	if err := s.Edit(context.Background(), 1, teacher.Profile{Name: "张三", Department: "理学院", Fields: map[string]string{"办公地点": " 办公室 "}}); err != nil {
		t.Fatal(err)
	}
	s.status.State = "running"
	if err := s.Edit(context.Background(), 1, teacher.Profile{}); err == nil {
		t.Fatal("editing during sync accepted")
	}
}
