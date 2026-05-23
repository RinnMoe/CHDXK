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

func TestUsernameFromEmail(t *testing.T) {
	deriver := NewBLAKE2bUsernameDeriver("SALT")
	got, err := deriver.UsernameFromEmail("Alice@Example.EDU")
	if err != nil {
		t.Fatalf("UsernameFromEmail: %v", err)
	}
	if got != "ca96242b3be38504f62aedd9191470aa" {
		t.Fatalf("UsernameFromEmail = %q", got)
	}

	if _, err := deriver.UsernameFromEmail("not an email"); err == nil {
		t.Fatal("expected invalid email error")
	}
}
