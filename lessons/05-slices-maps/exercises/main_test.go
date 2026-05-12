package exercises

import (
	"errors"
	"reflect"
	"testing"
)

// TestTotalsByCategory is a SKELETON. Fill in the cases and the t.Run body.
//
// Cases to cover:
//   - a basic happy-path case (no duplicate categories)
//   - a case where one category appears multiple times (totals accumulate)
//   - a single-element case
//   - the two error cases: empty inputs AND length mismatch
//
// For the error cases: the wantErr tag tells you which sentinel to expect.
// Use errors.Is to compare. The `want` map should be nil for error cases.
type totalsCase struct {
	name       string
	amounts    []float64
	categories []string
	want       map[string]float64
	wantErr    error // nil, errEmptyInputs, or errLengthMismatch
}

func TestTotalsByCategory(t *testing.T) {
	cases := []totalsCase{
		// TODO: add at least 5 cases here:
		//   1. Basic distinct categories
		//   2. Repeated category (totals accumulate)
		//   3. Single element
		//   4. Empty inputs → wantErr: errEmptyInputs
		//   5. Length mismatch → wantErr: errLengthMismatch
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := TotalsByCategory(tc.amounts, tc.categories)
			//   If tc.wantErr != nil: assert err is non-nil AND errors.Is(err, tc.wantErr).
			//   Otherwise: assert err == nil AND reflect.DeepEqual(got, tc.want).
			_ = tc
			_ = errors.Is
			_ = reflect.DeepEqual
		})
	}
}
