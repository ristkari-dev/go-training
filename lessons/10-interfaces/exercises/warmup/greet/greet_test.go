package greet

import (
	"strings"
	"testing"
)

// TestGreeters is a SKELETON. The point of the test is to demonstrate
// that English and Finnish BOTH satisfy the Greeter interface — without
// either type declaring that it does. We loop over `[]Greeter{...}` and
// call Greet on each.
//
// Cases to add: at minimum, English with "World" → "Hello, World!", and
// Finnish with "World" → "Hei, World!". Use strings.HasPrefix for a
// language-agnostic shape check, OR direct equality for exact match —
// your choice.
func TestGreeters(t *testing.T) {
	cases := []struct {
		name string
		g    Greeter
		in   string
		want string
	}{
		// TODO: at least 4 cases. English with "World", English with
		// "Aki", Finnish with "World", Finnish with empty string.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.g.Greet(tc.in)
			//   if got != tc.want { t.Errorf(...) }
			_ = tc
			_ = strings.HasPrefix
		})
	}
}

// TestGreeterSlice is a SKELETON. Verify that a `[]Greeter` slice
// containing both an English and a Finnish value works as one
// homogeneous collection — that's the payoff of implicit satisfaction.
func TestGreeterSlice(t *testing.T) {
	// TODO:
	//   greeters := []Greeter{English{}, Finnish{}}
	//   for _, g := range greeters {
	//       got := g.Greet("World")
	//       if got == "" { t.Error("empty greeting") }
	//   }
	_ = t
}
