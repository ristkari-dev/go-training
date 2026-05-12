package solutions

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
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Categorise(tc.amount); got != tc.want {
				t.Errorf("Categorise(%g) = %q, want %q", tc.amount, got, tc.want)
			}
		})
	}
}

func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name   string
		date   string
		amount float64
		cat    string
		want   string
	}{
		{"coffee", "2026-05-12", 4.50, "coffee", "2026-05-12  €4.50    coffee"},
		{"lunch", "2026-05-12", 12.00, "lunch", "2026-05-12  €12.00   lunch"},
		{"rent", "2026-05-12", 999.99, "rent", "2026-05-12  €999.99  rent"},
		{"zero-amount", "2026-05-12", 0, "free", "2026-05-12  €0.00    free"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExpense(tc.date, tc.amount, tc.cat)
			if got != tc.want {
				t.Errorf("FormatExpense(%q, %g, %q) =\n  %q\nwant\n  %q", tc.date, tc.amount, tc.cat, got, tc.want)
			}
		})
	}
}
