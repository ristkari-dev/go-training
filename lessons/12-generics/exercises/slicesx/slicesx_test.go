package slicesx

import (
	"strconv"
	"strings"
	"testing"
)

// TestFilterInts is a SKELETON. Cover at least:
//   - Filter even numbers from a mix of positives/negatives
//   - Filter with a predicate that matches nothing → empty (or nil)
//   - Filter empty input → empty
func TestFilterInts(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	cases := []struct {
		name string
		in   []int
		pred func(int) bool
		want []int
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := Filter(tc.in, tc.pred); compare to tc.want.
			_ = tc
			_ = even
		})
	}
}

// TestFilterStrings is a SKELETON. Filter strings by length (e.g., keep
// only strings of length > 1). Demonstrates that the same Filter works
// with any element type — type inference picks T=string.
func TestFilterStrings(t *testing.T) {
	longerThan1 := func(s string) bool { return len(s) > 1 }
	// TODO: write one assertion. E.g.:
	//   got := Filter([]string{"a", "bb", "ccc"}, longerThan1)
	//   want := []string{"bb", "ccc"}
	//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
	_ = longerThan1
}

// TestMapIntToString is a SKELETON. Demonstrates Map with T=int, U=string.
// Use strconv.Itoa as the transform.
func TestMapIntToString(t *testing.T) {
	// TODO:
	//   got := Map([]int{1, 2, 3}, strconv.Itoa)
	//   want := []string{"1", "2", "3"}
	//   if !reflect.DeepEqual(got, want) → t.Errorf("...")
	_ = strconv.Itoa
}

// TestMapStringToString is a SKELETON. Demonstrates Map with T=U=string.
// Use strings.ToUpper as the transform.
func TestMapStringToString(t *testing.T) {
	// TODO:
	//   got := Map([]string{"hi", "world"}, strings.ToUpper)
	//   want := []string{"HI", "WORLD"}
	_ = strings.ToUpper
}
