package exercises

import (
	"errors"
	"testing"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative-cancels", -1, 1, 0},
		{"both-negative", -5, -7, -12},
		{"zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Add(tc.a, tc.b); got != tc.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	cases := []struct {
		name             string
		xs               []int
		wantMin, wantMax int
		wantErr          bool
	}{
		{"single", []int{42}, 42, 42, false},
		{"sorted", []int{1, 2, 3, 4, 5}, 1, 5, false},
		{"reverse", []int{5, 4, 3, 2, 1}, 1, 5, false},
		{"unsorted", []int{3, 1, 4, 1, 5, 9, 2, 6}, 1, 9, false},
		{"all-negative", []int{-3, -1, -7, -2}, -7, -1, false},
		{"mixed-signs", []int{-5, 0, 5}, -5, 5, false},
		{"empty", []int{}, 0, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMin, gotMax, err := MinMax(tc.xs...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("MinMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyMinMax) {
					t.Errorf("MinMax(%v) error = %v, want errEmptyMinMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("MinMax(%v) unexpected error: %v", tc.xs, err)
			}
			if gotMin != tc.wantMin {
				t.Errorf("MinMax(%v) min = %d, want %d", tc.xs, gotMin, tc.wantMin)
			}
			if gotMax != tc.wantMax {
				t.Errorf("MinMax(%v) max = %d, want %d", tc.xs, gotMax, tc.wantMax)
			}
		})
	}
}
