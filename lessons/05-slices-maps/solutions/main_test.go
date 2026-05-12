package solutions

import (
	"errors"
	"reflect"
	"testing"
)

type totalsCase struct {
	name       string
	amounts    []float64
	categories []string
	want       map[string]float64
	wantErr    error
}

func TestTotalsByCategory(t *testing.T) {
	cases := []totalsCase{
		{
			"distinct",
			[]float64{4.50, 12, 75},
			[]string{"coffee", "lunch", "rent"},
			map[string]float64{"coffee": 4.50, "lunch": 12, "rent": 75},
			nil,
		},
		{
			"repeated-categories",
			[]float64{4.50, 12, 75, 9.99},
			[]string{"coffee", "lunch", "rent", "coffee"},
			map[string]float64{"coffee": 14.49, "lunch": 12, "rent": 75},
			nil,
		},
		{
			"all-same-category",
			[]float64{1, 2, 3},
			[]string{"x", "x", "x"},
			map[string]float64{"x": 6},
			nil,
		},
		{
			"single",
			[]float64{42},
			[]string{"a"},
			map[string]float64{"a": 42},
			nil,
		},
		{
			"empty-amounts",
			[]float64{},
			[]string{"a"},
			nil,
			errEmptyInputs,
		},
		{
			"empty-both",
			[]float64{},
			[]string{},
			nil,
			errEmptyInputs,
		},
		{
			"nil-both",
			nil,
			nil,
			nil,
			errEmptyInputs,
		},
		{
			"length-mismatch-short-categories",
			[]float64{1, 2, 3},
			[]string{"a", "b"},
			nil,
			errLengthMismatch,
		},
		{
			"length-mismatch-short-amounts",
			[]float64{1, 2},
			[]string{"a", "b", "c"},
			nil,
			errLengthMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TotalsByCategory(tc.amounts, tc.categories)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("TotalsByCategory: nil error, want %v", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("TotalsByCategory: err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("TotalsByCategory: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("TotalsByCategory: got %v, want %v", got, tc.want)
			}
		})
	}
}
