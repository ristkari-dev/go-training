package summary

import (
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/exercises/internal/expense"
)

func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []expense.Expense
		want map[string]float64
	}{
		{
			"distinct",
			[]expense.Expense{
				{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-18", Amount: 12, Category: "lunch"},
				{Date: "2026-05-18", Amount: 75, Category: "rent"},
			},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
		},
		{
			"repeated",
			[]expense.Expense{
				{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
				{Date: "2026-05-18", Amount: 12, Category: "lunch"},
				{Date: "2026-05-18", Amount: 9.99, Category: "coffee"},
			},
			map[string]float64{"coffee": 14.49, "lunch": 12},
		},
		{"single", []expense.Expense{{Amount: 42, Category: "x"}}, map[string]float64{"x": 42}},
		{"empty", []expense.Expense{}, map[string]float64{}},
		{"nil", nil, map[string]float64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TotalsByCategory(tc.es)
			if got == nil {
				t.Fatalf("TotalsByCategory returned nil; expected non-nil map")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBiggestCategory(t *testing.T) {
	cases := []struct {
		name      string
		totals    map[string]float64
		wantName  string
		wantTotal float64
	}{
		{"single", map[string]float64{"a": 10}, "a", 10},
		{"clear-max", map[string]float64{"coffee": 14.49, "rent": 75, "lunch": 12}, "rent", 75},
		{"tie-alphabetical", map[string]float64{"a": 10, "b": 10}, "a", 10},
		{"tie-three-way", map[string]float64{"c": 5, "b": 5, "a": 5}, "a", 5},
		{"empty", map[string]float64{}, "", 0},
		{"nil", nil, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotTotal := BiggestCategory(tc.totals)
			if gotName != tc.wantName || gotTotal != tc.wantTotal {
				t.Errorf("BiggestCategory(%v) = (%q, %g), want (%q, %g)",
					tc.totals, gotName, gotTotal, tc.wantName, tc.wantTotal)
			}
		})
	}
}

func TestBar(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		max   float64
		width int
		want  string
	}{
		{"full", 10, 10, 8, "████████"},
		{"empty-zero-value", 0, 10, 8, ""},
		{"empty-zero-max", 5, 0, 8, ""},
		{"empty-zero-width", 5, 10, 0, ""},
		{"negative-value", -1, 10, 8, ""},
		{"half-units-rounds-down", 5, 10, 8, "████"},
		{"quarter-units", 2.5, 10, 8, "██"},
		{"half-block-only", 0.5, 8, 8, "▌"},
		{"one-full-block", 1, 8, 8, "█"},
		{"one-and-a-half", 1.5, 8, 8, "█▌"},
		{"seven-and-a-half", 7.5, 8, 8, "███████▌"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Bar(tc.value, tc.max, tc.width); got != tc.want {
				t.Errorf("Bar(%g, %g, %d) = %q, want %q", tc.value, tc.max, tc.width, got, tc.want)
			}
		})
	}
}
