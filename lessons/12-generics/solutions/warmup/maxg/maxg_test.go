package maxg

import (
	"testing"
)

func TestMaxInts(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want int
	}{
		{"positive", []int{1, 5, 3, 2}, 5},
		{"negative", []int{-1, -5, -3}, -1},
		{"mixed", []int{-2, 0, 7, -100, 3}, 7},
		{"single", []int{42}, 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.in)
			if err != nil {
				t.Fatalf("Max(%v) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Max(%v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestMaxStrings(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"fruits", []string{"apple", "banana", "cherry"}, "cherry"},
		{"single", []string{"only"}, "only"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Max(tc.in)
			if err != nil {
				t.Fatalf("Max(%v) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Max(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMaxEmpty(t *testing.T) {
	if _, err := Max([]int{}); err == nil {
		t.Fatal("Max([]int{}): expected error, got nil")
	}
	// nil slice — type inference can't see []int from nil alone, so we
	// have to be explicit. (This is also one of the "Common mistake"
	// slide examples in concept 3 / type inference.)
	if _, err := Max[int](nil); err == nil {
		t.Fatal("Max[int](nil): expected error, got nil")
	}
}
