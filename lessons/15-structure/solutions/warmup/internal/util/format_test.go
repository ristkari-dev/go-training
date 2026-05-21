package util

import "testing"

func TestMoney(t *testing.T) {
	cases := []struct {
		name string
		in   float64
		want string
	}{
		{"whole-number", 4.0, "€4.00"},
		{"two-decimals", 4.50, "€4.50"},
		{"three-decimals-rounded-up", 4.567, "€4.57"},
		{"zero", 0.0, "€0.00"},
		{"large", 12345.67, "€12345.67"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Money(tc.in)
			if got != tc.want {
				t.Errorf("Money(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
