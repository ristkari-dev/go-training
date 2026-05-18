package expense

import "testing"

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-18", Amount: 4.50, Category: "coffee"}, "2026-05-18  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-18", Amount: 12, Category: "lunch"}, "2026-05-18  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-18", Amount: 999.99, Category: "rent"}, "2026-05-18  €999.99  rent"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.Format(); got != tc.want {
				t.Errorf("%+v.Format() =\n  %q\nwant\n  %q", tc.e, got, tc.want)
			}
		})
	}
}

func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		{"zero", 0, false},
		{"small", 4.50, false},
		{"just-under-50", 49.99, false},
		{"exactly-50", 50, false},
		{"just-over-50", 50.01, true},
		{"large", 999, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := (Expense{Amount: tc.amount}).IsHigh(); got != tc.want {
				t.Errorf("Expense{Amount: %g}.IsHigh() = %v, want %v", tc.amount, got, tc.want)
			}
		})
	}
}
