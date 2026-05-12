package exercises

import (
	"errors"
	"reflect"
	"testing"
)

// TestWarmupSum is a SKELETON. The struct and the loop scaffold are in
// place; you fill in the cases and the t.Run body.
//
// Cases to cover: a basic positive case, a single-element case, an
// empty slice, a nil slice, and a case mixing positive and negative.
func TestWarmupSum(t *testing.T) {
	cases := []struct {
		name string
		xs   []int
		want int
	}{
		// TODO: add at least 4 cases. Use the examples in the WarmupSum
		// doc comment as a starting point.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: call WarmupSum(tc.xs), compare with tc.want, t.Errorf if differ.
			_ = tc
		})
	}
}

// TestWarmupMax is a SKELETON. Similar shape to TestWarmupSum, but the
// function returns (int, error) — your assertion needs to branch on the
// wantErr flag.
//
// Cases to cover: a basic case, a single-element case, all-negative,
// AND the empty-input case (which should return errEmptyWarmupMax).
func TestWarmupMax(t *testing.T) {
	cases := []struct {
		name    string
		xs      []int
		want    int
		wantErr bool
	}{
		// TODO: at least 4 cases. Include the empty case (wantErr: true).
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: gotMax, err := WarmupMax(tc.xs)
			// If tc.wantErr: assert err is non-nil AND errors.Is(err, errEmptyWarmupMax).
			// Otherwise: assert err == nil AND gotMax == tc.want.
			_ = tc
			_ = errors.Is // satisfy the import while you write the test
		})
	}
}

// TestWarmupUnique is a SKELETON. Compare slices with reflect.DeepEqual.
// Be careful with nil vs empty: WarmupUnique always returns a non-nil slice.
func TestWarmupUnique(t *testing.T) {
	cases := []struct {
		name string
		xs   []string
		want []string
	}{
		// TODO: at least 4 cases. Cover duplicates-with-order-preserved,
		// a single-element case, an empty input, and a nil input.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := WarmupUnique(tc.xs)
			// Assert got != nil (it must be a non-nil empty slice for empty/nil input).
			// Assert reflect.DeepEqual(got, tc.want) is true.
			_ = tc
			_ = reflect.DeepEqual // satisfy the import while you write the test
		})
	}
}
