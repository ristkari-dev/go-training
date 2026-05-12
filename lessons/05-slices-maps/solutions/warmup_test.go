package solutions

import (
	"errors"
	"reflect"
	"testing"
)

func TestWarmupSum(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
		want int
	}{
		{"basic", []int{1, 2, 3}, 6},
		{"single", []int{42}, 42},
		{"mixed-signs", []int{-3, 5, -7, 2}, -3},
		{"all-negative", []int{-1, -2, -3}, -6},
		{"empty", []int{}, 0},
		{"nil", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WarmupSum(tc.xs); got != tc.want {
				t.Errorf("WarmupSum(%v) = %d, want %d", tc.xs, got, tc.want)
			}
		})
	}
}

func TestWarmupMax(t *testing.T) {
	cases := []struct {
		name    string
		xs      []int
		want    int
		wantErr bool
	}{
		{"unsorted", []int{3, 1, 4, 1, 5, 9, 2, 6}, 9, false},
		{"sorted-ascending", []int{1, 2, 3, 4, 5}, 5, false},
		{"sorted-descending", []int{5, 4, 3, 2, 1}, 5, false},
		{"single", []int{42}, 42, false},
		{"all-negative", []int{-3, -1, -7}, -1, false},
		{"empty", []int{}, 0, true},
		{"nil", nil, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := WarmupMax(tc.xs)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("WarmupMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyWarmupMax) {
					t.Errorf("WarmupMax(%v) error = %v, want errEmptyWarmupMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("WarmupMax(%v) unexpected error: %v", tc.xs, err)
			}
			if got != tc.want {
				t.Errorf("WarmupMax(%v) = %d, want %d", tc.xs, got, tc.want)
			}
		})
	}
}

func TestWarmupUnique(t *testing.T) {
	cases := []struct {
		name string
		xs   []string
		want []string
	}{
		{"basic-dedupe", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"no-dupes", []string{"x", "y", "z"}, []string{"x", "y", "z"}},
		{"all-same", []string{"x", "x", "x"}, []string{"x"}},
		{"single", []string{"x"}, []string{"x"}},
		{"empty", []string{}, []string{}},
		{"nil", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupUnique(tc.xs)
			if got == nil {
				t.Fatalf("WarmupUnique(%v) returned nil; expected a non-nil slice", tc.xs)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("WarmupUnique(%v) = %v, want %v", tc.xs, got, tc.want)
			}
		})
	}
}
