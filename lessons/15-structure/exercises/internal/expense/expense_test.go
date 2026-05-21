package expense

import "testing"

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-19", Amount: 4.50, Category: "coffee"}, "2026-05-19  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-19", Amount: 12, Category: "lunch"}, "2026-05-19  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-19", Amount: 999.99, Category: "rent"}, "2026-05-19  €999.99  rent"},
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

func TestExpenseApplyDiscount(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		rate        float64
		wantAmount  float64
	}{
		{"10-percent-off-100", 100, 0.10, 90},
		{"50-percent-off-50", 50, 0.50, 25},
		{"zero-rate-no-op", 50, 0, 50},
		{"full-discount", 100, 1.0, 0},
		{"discount-on-zero", 0, 0.10, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Expense{Amount: tc.startAmount}
			e.ApplyDiscount(tc.rate)
			if e.Amount != tc.wantAmount {
				t.Errorf("Expense{Amount: %g}.ApplyDiscount(%g) → Amount %g, want %g",
					tc.startAmount, tc.rate, e.Amount, tc.wantAmount)
			}
		})
	}
}

func TestExpenseBump(t *testing.T) {
	cases := []struct {
		name        string
		startAmount float64
		bump        float64
		wantAmount  float64
	}{
		{"positive-bump", 10, 5, 15},
		{"negative-bump", 10, -3, 7},
		{"zero-bump", 10, 0, 10},
		{"bump-on-zero", 0, 5, 5},
		{"bump-to-negative", 5, -10, -5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := Expense{Amount: tc.startAmount}
			e.Bump(tc.bump)
			if e.Amount != tc.wantAmount {
				t.Errorf("Expense{Amount: %g}.Bump(%g) → Amount %g, want %g",
					tc.startAmount, tc.bump, e.Amount, tc.wantAmount)
			}
		})
	}
}
