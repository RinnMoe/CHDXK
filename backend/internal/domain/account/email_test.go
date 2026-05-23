package account

import "testing"

func TestEmailWhitelist_AllowsExactAndDomain(t *testing.T) {
	w := NewEmailWhitelist([]string{"alice@example.edu", "@jcourse.edu"})

	cases := []struct {
		email string
		want  bool
	}{
		{email: "alice@example.edu", want: true},
		{email: "bob@jcourse.edu", want: true},
		{email: "bob@example.com", want: false},
	}

	for _, tc := range cases {
		if got := w.Allows(tc.email); got != tc.want {
			t.Fatalf("Allows(%q) = %v, want %v", tc.email, got, tc.want)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail("Alice@Example.EDU")
	if err != nil {
		t.Fatalf("NormalizeEmail: %v", err)
	}
	if got != "alice@example.edu" {
		t.Fatalf("NormalizeEmail = %q", got)
	}

	if _, err := NormalizeEmail("not an email"); err == nil {
		t.Fatal("expected invalid email error")
	}
}
