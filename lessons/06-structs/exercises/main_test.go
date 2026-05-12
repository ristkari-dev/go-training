package exercises

import (
	"reflect"
	"testing"
)

// TestExpenseFormat is a SKELETON. Fill in cases and the t.Run body.
//
// Build the expected output strings by hand using "%s  €%-7.2f %s".
// Use the examples in the Format doc comment as a starting point.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases. Include the small/medium/large amounts
		// from the doc comment so you exercise the %-7.2f width on values
		// of different magnitude.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover the snack/regular/splurge-like
// boundaries: amount well under 50, exactly 50 (NOT high), just over 50,
// and a large value.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases including the boundary at 50.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Expense{Amount: tc.amount}.IsHigh()
			//   compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}

// TestTotalsByCategory is a SKELETON. Cover at least:
//   - distinct categories (no overlap)
//   - repeated categories (totals must accumulate)
//   - a single-element slice
//   - an empty slice (result is a non-nil empty map)
//   - nil input (also returns an empty map)
func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []Expense
		want map[string]float64
	}{
		// TODO: at least 4 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := TotalsByCategory(tc.es)
			//   assert got != nil (always returns a non-nil map)
			//   assert reflect.DeepEqual(got, tc.want)
			_ = tc
			_ = reflect.DeepEqual
		})
	}
}
