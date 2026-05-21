package util

import "testing"

// TestMoney is a SKELETON. Assert that Money returns the expected
// "€<2-decimal>" formatting for at least 3 inputs.
func TestMoney(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		// TODO: at least 3 cases.
		// {"whole-number", 4.0, "€4.00"},
		// {"two-decimals", 4.50, "€4.50"},
		// {"long-decimals-rounded", 4.567, "€4.57"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got := Money(tc.in)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}
