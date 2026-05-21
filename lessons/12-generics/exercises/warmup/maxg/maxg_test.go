package maxg

import (
	"testing"
)

// TestMaxInts is a SKELETON. Cover at least:
//   - A slice of positive ints (e.g., {1, 5, 3, 2} → 5)
//   - A slice of negative ints (e.g., {-1, -5, -3} → -1)
//   - A single-element slice (e.g., {42} → 42)
func TestMaxInts(t *testing.T) {
	cases := []struct {
		name string
		in   []int
		want int
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := Max(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestMaxStrings is a SKELETON. Demonstrate that Max works on []string
// too — same call, different element type (compiler infers T=string).
func TestMaxStrings(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		// TODO: at least 2 cases.
		// Hint: strings compare lexicographically. "banana" > "apple".
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got, err := Max(tc.in); assert.
			_ = tc
		})
	}
}

// TestMaxEmpty is a SKELETON. Assert that Max(nil) and Max([]int{}) both
// return a non-nil error (and a zero value, but the zero matters less —
// the error is the signal).
func TestMaxEmpty(t *testing.T) {
	// TODO:
	//   _, err := Max([]int{})
	//   if err == nil → t.Fatal("expected error for empty slice")
	//   _, err = Max[int](nil)  // explicit type param needed; nil has no element type
	//   if err == nil → t.Fatal("expected error for nil slice")
}
