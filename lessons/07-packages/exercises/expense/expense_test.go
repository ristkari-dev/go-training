package expense

import (
	"reflect"
	"testing"
)

// TestExpenseFormat is a SKELETON. Fill in cases and the t.Run body.
// Use the doc-comment examples in expense.go as a starting point.
func TestExpenseFormat(t *testing.T) {
	cases := []struct {
		name string
		e    Expense
		want string
	}{
		// TODO: at least 3 cases. Cover small/medium/large amounts so
		// you exercise the %-7.2f width on different magnitudes.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.e.Format(); compare with tc.want; t.Errorf if differ.
			_ = tc
		})
	}
}

// TestExpenseIsHigh is a SKELETON. Cover the boundary at 50 explicitly.
func TestExpenseIsHigh(t *testing.T) {
	cases := []struct {
		name   string
		amount float64
		want   bool
	}{
		// TODO: at least 4 cases including 49.99, 50, 50.01.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Expense{Amount: tc.amount}.IsHigh()
			_ = tc
		})
	}
}

// TestTotalsByCategory is a SKELETON. Cover at least:
//   - distinct categories
//   - repeated categories (totals accumulate)
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
