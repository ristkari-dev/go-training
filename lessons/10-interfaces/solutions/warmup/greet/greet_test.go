package greet

import "testing"

func TestGreeters(t *testing.T) {
	cases := []struct {
		name string
		g    Greeter
		in   string
		want string
	}{
		{"english-world", English{}, "World", "Hello, World!"},
		{"english-name", English{}, "Aki", "Hello, Aki!"},
		{"finnish-world", Finnish{}, "World", "Hei, World!"},
		{"finnish-name", Finnish{}, "Aki", "Hei, Aki!"},
		{"english-empty", English{}, "", "Hello, !"},
		{"finnish-empty", Finnish{}, "", "Hei, !"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.g.Greet(tc.in); got != tc.want {
				t.Errorf("%T.Greet(%q) = %q, want %q", tc.g, tc.in, got, tc.want)
			}
		})
	}
}

func TestGreeterSlice(t *testing.T) {
	greeters := []Greeter{English{}, Finnish{}}
	for _, g := range greeters {
		got := g.Greet("World")
		if got == "" {
			t.Errorf("%T.Greet returned empty string", g)
		}
	}
}
