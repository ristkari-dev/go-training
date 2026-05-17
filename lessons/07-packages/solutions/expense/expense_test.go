package expense

import (
	"reflect"
	"testing"
)

func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		{"coffee", Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}, "2026-05-15  €4.50    coffee"},
		{"lunch", Expense{Date: "2026-05-15", Amount: 12, Category: "lunch"}, "2026-05-15  €12.00   lunch"},
		{"rent", Expense{Date: "2026-05-15", Amount: 999.99, Category: "rent"}, "2026-05-15  €999.99  rent"},
		{"zero-amount", Expense{Date: "2026-05-15", Amount: 0, Category: "free"}, "2026-05-15  €0.00    free"},
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

func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []Expense
		want map[string]float64
	}{
		{
			"distinct",
			[]Expense{
				{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-15", Amount: 12, Category: "lunch"},
				{Date: "2026-05-15", Amount: 75, Category: "rent"},
			},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
		},
		{
			"repeated-categories",
			[]Expense{
				{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-15", Amount: 12, Category: "lunch"},
				{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
			},
			map[string]float64{"coffee": 14.49, "lunch": 12},
		},
		{
			"single",
			[]Expense{{Date: "2026-05-15", Amount: 42, Category: "x"}},
			map[string]float64{"x": 42},
		},
		{"empty", []Expense{}, map[string]float64{}},
		{"nil", nil, map[string]float64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TotalsByCategory(tc.es)
			if got == nil {
				t.Fatalf("TotalsByCategory(%v) returned nil; expected a non-nil map", tc.es)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory(%v) = %v, want %v", tc.es, got, tc.want)
			}
		})
	}
}
