package solutions

import "testing"

func TestWarmupHello(t *testing.T) {
	got := WarmupHello()
	want := "Hello, Go!"
	if got != want {
		t.Errorf("WarmupHello() = %q, want %q", got, want)
	}
}

func TestWarmupGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"World", "World", "Hello, World!"},
		{"Aki", "Aki", "Hello, Aki!"},
		{"empty", "", "Hello, !"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupGreet(tc.in)
			if got != tc.want {
				t.Errorf("WarmupGreet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
