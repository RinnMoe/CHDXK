package canteen

import (
	"reflect"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "trim spaces", in: "  麻辣烫  ", want: "麻辣烫"},
		{name: "drop punctuation", in: "麻辣烫！", want: "麻辣烫"},
		{name: "full width to half width", in: "ＡＢＣ", want: "abc"},
		{name: "lowercase and separators", in: "A B-C", want: "abc"},
		{name: "empty", in: "   ", want: ""},
		{name: "symbols dropped", in: "面馆·二层", want: "面馆二层"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeName(tc.in); got != tc.want {
				t.Fatalf("NormalizeName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeNameEquivalence(t *testing.T) {
	variants := []string{"麻辣烫", " 麻辣烫 ", "麻辣烫!", "麻辣烫！", "麻 辣 烫"}
	want := NormalizeName(variants[0])
	for _, variant := range variants[1:] {
		if got := NormalizeName(variant); got != want {
			t.Fatalf("NormalizeName(%q) = %q, want %q", variant, got, want)
		}
	}
}

func TestNormalizeCategories(t *testing.T) {
	got := NormalizeCategories([]string{" 面食 ", "面食", "", "快餐", "快餐", "汤类"})
	want := []string{"面食", "快餐", "汤类"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NormalizeCategories = %v, want %v", got, want)
	}

	tooMany := make([]string, 0, MaxStoreCategories+2)
	for _, category := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		tooMany = append(tooMany, category)
	}
	if got := NormalizeCategories(tooMany); len(got) != MaxStoreCategories {
		t.Fatalf("NormalizeCategories length = %d, want %d", len(got), MaxStoreCategories)
	}
}
