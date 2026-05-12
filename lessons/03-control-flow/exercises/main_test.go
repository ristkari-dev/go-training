package exercises

import "testing"

func TestGreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"World", "Hello, World!"},
		{"Go", "Hello, Go!"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := Greet(tc.in)
			if got != tc.want {
				t.Errorf("Greet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
