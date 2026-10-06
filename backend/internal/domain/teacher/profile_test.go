package teacher

import "testing"

func TestProfileCandidatesRequiresUniqueNameAndCollege(t *testing.T) {
	a := TeacherView{ID: 1, Name: "张三", Department: "理学院"}
	b := TeacherView{ID: 2, Name: "张三", Department: "信息工程学院"}
	profiles := []Profile{{URL: "a", Name: "张三", Department: "理学院"}, {URL: "b", Name: "张三", Department: "信息工程学院"}}
	candidates, unique := ProfileCandidates(a, []TeacherView{a, b}, profiles)
	if !unique || len(candidates) != 1 || candidates[0].URL != "a" {
		t.Fatal("did not disambiguate college")
	}
	_, unique = ProfileCandidates(a, []TeacherView{a, a}, profiles)
	if unique {
		t.Fatal("local duplicates automatically merged")
	}
	profiles = append(profiles, Profile{URL: "c", Name: "张三", Department: "理学院"})
	_, unique = ProfileCandidates(a, []TeacherView{a, b}, profiles)
	if unique {
		t.Fatal("source duplicates automatically merged")
	}
	candidates, unique = ProfileCandidates(TeacherView{Name: "李四", Department: "理学院"}, []TeacherView{a, b}, profiles)
	if unique || len(candidates) != 0 {
		t.Fatal("missing teacher matched")
	}
	if ProfileIdentity(" 张 三 ", "某学院（系）") != ProfileIdentity("张三", "某学院(系)") {
		t.Fatal("equivalent whitespace or punctuation mismatched")
	}
}
