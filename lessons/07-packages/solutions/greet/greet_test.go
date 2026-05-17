package greet

import "testing"

func TestGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"world", "World", "Hello, World!"},
		{"name", "Aki", "Hello, Aki!"},
		{"empty", "", "Hello, !"},
		{"unicode", "Étienne", "Hello, Étienne!"},
		{"with-space", "Go Course", "Hello, Go Course!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Greet(tc.in); got != tc.want {
				t.Errorf("Greet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
