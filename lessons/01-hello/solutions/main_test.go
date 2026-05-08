package solutions

import "testing"

func TestGreet(t *testing.T) {
	got := Greet("World")
	want := "Hello, World!"
	if got != want {
		t.Errorf("Greet(%q) = %q, want %q", "World", got, want)
	}
}

func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name     string
		date     string
		amount   float64
		category string
		want     string
	}{
		{"coffee", "2026-05-07", 4.50, "coffee", "2026-05-07  €4.50  coffee"},
		{"groceries", "2026-05-07", 23.5, "groceries", "2026-05-07  €23.50  groceries"},
		{"big-rent", "2026-04-30", 1234.5, "rent", "2026-04-30  €1234.50  rent"},
		{"zero", "2026-05-07", 0, "free", "2026-05-07  €0.00  free"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExpense(tc.date, tc.amount, tc.category)
			if got != tc.want {
				t.Errorf("FormatExpense(%q, %g, %q) = %q, want %q",
					tc.date, tc.amount, tc.category, got, tc.want)
			}
		})
	}
}
