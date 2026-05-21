package parseage

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

// TestParseAgeValid is a SKELETON. Cover at least two cases where the
// input is a valid integer (positive, zero) and assert (age, nil err).
func TestParseAgeValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		// TODO: at least 3 cases. "25" → 25, "0" → 0, "-1" → -1, etc.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got, err := ParseAge(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseAgeInvalidWraps is a SKELETON. For invalid inputs, assert
// THREE things:
//  1. The error is non-nil.
//  2. The error MESSAGE contains the offending input (so humans can see
//     what was wrong). Use strings.Contains.
//  3. errors.Is(err, strconv.ErrSyntax) is true — the underlying
//     strconv sentinel survives the wrap.
//
// Cases: empty string, non-numeric "twenty", spaces "  25  ".
func TestParseAgeInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   _, err := ParseAge(tc.in)
			//   if err == nil → t.Fatalf("expected error for %q", tc.in)
			//   if !strings.Contains(err.Error(), tc.in) → t.Errorf("message missing input")
			//   if !errors.Is(err, strconv.ErrSyntax) → t.Errorf("wrapped sentinel not findable")
			_ = tc
			_ = errors.Is
			_ = strconv.ErrSyntax
			_ = strings.Contains
		})
	}
}
