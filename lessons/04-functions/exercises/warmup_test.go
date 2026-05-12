package exercises

import (
	"errors"
	"testing"
)

func TestWarmupAdd(t *testing.T) {
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
			if got := WarmupAdd(tc.a, tc.b); got != tc.want {
				t.Errorf("WarmupAdd(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestWarmupMinMax(t *testing.T) {
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
			gotMin, gotMax, err := WarmupMinMax(tc.xs...)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("WarmupMinMax(%v) returned nil error, want non-nil", tc.xs)
				}
				if !errors.Is(err, errEmptyWarmupMinMax) {
					t.Errorf("WarmupMinMax(%v) error = %v, want errEmptyWarmupMinMax", tc.xs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("WarmupMinMax(%v) unexpected error: %v", tc.xs, err)
			}
			if gotMin != tc.wantMin {
				t.Errorf("WarmupMinMax(%v) min = %d, want %d", tc.xs, gotMin, tc.wantMin)
			}
			if gotMax != tc.wantMax {
				t.Errorf("WarmupMinMax(%v) max = %d, want %d", tc.xs, gotMax, tc.wantMax)
			}
		})
	}
}
