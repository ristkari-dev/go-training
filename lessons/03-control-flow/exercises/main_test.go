package exercises

import "testing"

func TestCategorise(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   string
	}{
		{"zero-snack", 0, "snack"},
		{"small-snack", 4.50, "snack"},
		{"upper-snack-edge", 9.99, "snack"},
		{"lower-regular-edge", 10.00, "regular"},
		{"middle-regular", 12.00, "regular"},
		{"upper-regular-edge", 50.00, "regular"},
		{"just-above-regular", 50.01, "splurge"},
		{"large-splurge", 75.00, "splurge"},
		{"very-large-splurge", 999.99, "splurge"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Categorise(tc.amount); got != tc.want {
				t.Errorf("Categorise(%g) = %q, want %q", tc.amount, got, tc.want)
			}
		})
	}
}

func TestTally(t *testing.T) {
	cases := []struct {
		name                            string
		amounts                         []float64
		wantSnack, wantRegular, wantSpl int
	}{
		{"empty", []float64{}, 0, 0, 0},
		{"all-snack", []float64{1, 2, 9.99}, 3, 0, 0},
		{"all-regular", []float64{10, 25, 50}, 0, 3, 0},
		{"all-splurge", []float64{75, 100, 999}, 0, 0, 3},
		{"mixed", []float64{4.50, 12, 75, 9.99}, 2, 1, 1},
		{"boundary-mix", []float64{9.99, 10, 50, 50.01}, 1, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snack, regular, splurge := Tally(tc.amounts)
			if snack != tc.wantSnack || regular != tc.wantRegular || splurge != tc.wantSpl {
				t.Errorf("Tally(%v) = (%d, %d, %d), want (%d, %d, %d)",
					tc.amounts, snack, regular, splurge,
					tc.wantSnack, tc.wantRegular, tc.wantSpl)
			}
		})
	}
}
