package greet

import "testing"

// TestGreet is a SKELETON. Fill in the cases and the t.Run body.
//
// Cases to cover: a normal name, the literal "World", an empty name,
// and at least one with non-ASCII characters (e.g. "Aki" or "Étienne").
func TestGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Greet(tc.in); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}
