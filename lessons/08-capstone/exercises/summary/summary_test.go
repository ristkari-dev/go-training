package summary

import (
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// TestTotalsByCategory is a SKELETON. Same shape as lesson 05/06/07's
// TotalsByCategory test.
func TestTotalsByCategory(t *testing.T) {
	cases := []struct {
		name string
		es   []expense.Expense
		want map[string]float64
	}{
		// TODO: at least 4 cases (distinct, repeated, single, empty).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := TotalsByCategory(tc.es)
			//   assert got != nil; reflect.DeepEqual(got, tc.want)
			_ = tc
			_ = reflect.DeepEqual
		})
	}
}

// TestBiggestCategory is a SKELETON. Cover at least:
//   - one category (trivial maximum)
//   - several categories with a clear maximum
//   - a tie (assert alphabetical tie-break)
//   - empty map (returns ("", 0))
//   - nil map (returns ("", 0))
func TestBiggestCategory(t *testing.T) {
	cases := []struct {
		name      string
		totals    map[string]float64
		wantName  string
		wantTotal float64
	}{
		// TODO: at least 5 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: gotName, gotTotal := BiggestCategory(tc.totals)
			//   assert gotName == tc.wantName && gotTotal == tc.wantTotal
			_ = tc
		})
	}
}

// TestBar is a SKELETON. Cover at least:
//   - full bar (value == max)
//   - empty bar (value == 0)
//   - half-block on the trailing position
//   - max <= 0 returns ""
//   - width <= 0 returns ""
func TestBar(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		max   float64
		width int
		want  string
	}{
		// TODO: at least 5 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Bar(tc.value, tc.max, tc.width)
			//   assert got == tc.want
			_ = tc
		})
	}
}
