package dates

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestFormatDate is a SKELETON. Build a few time.Time values via
// time.Date(year, month, day, hour, min, sec, nsec, time.UTC) and
// assert FormatDate produces the expected "YYYY-MM-DD HH:MM:SS" string.
func TestFormatDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		// TODO: at least 3 cases. E.g.:
		// {"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), "1970-01-01 00:00:00"},
		// {"sample", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC), "2026-05-21 14:30:00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := FormatDate(tc.in); if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseDateValid is a SKELETON. Pass valid date strings; assert
// ParseDate returns the expected time.Time + nil error.
func TestParseDateValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Time
	}{
		// TODO: at least 2 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   got, err := ParseDate(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if !got.Equal(tc.want) → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseDateInvalidWraps is a SKELETON. For invalid inputs, assert:
//   - err is non-nil
//   - err message contains the offending input
//   - errors.Unwrap returns a non-nil underlying time.Parse error
func TestParseDateInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		// TODO: at least 2 cases ("not-a-date", "2026-13-01 99:99:99", etc.)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   _, err := ParseDate(tc.in)
			//   if err == nil → t.Fatalf("expected error for %q", tc.in)
			//   if !strings.Contains(err.Error(), tc.in) → t.Errorf("error message missing input")
			//   if errors.Unwrap(err) == nil → t.Errorf("expected wrapped error")
			_ = tc
			_ = strings.Contains
			_ = errors.Unwrap
		})
	}
}

// TestRoundTrip is a SKELETON. ParseDate(FormatDate(t)) should equal t
// (truncated to seconds — Format drops sub-second precision).
func TestRoundTrip(t *testing.T) {
	// TODO:
	//   original := time.Date(2026, 5, 21, 14, 30, 45, 0, time.UTC)
	//   s := FormatDate(original)
	//   parsed, err := ParseDate(s)
	//   if err != nil → t.Fatalf("ParseDate: %v", err)
	//   if !parsed.Equal(original) → t.Errorf("roundtrip: %v != %v", parsed, original)
}
