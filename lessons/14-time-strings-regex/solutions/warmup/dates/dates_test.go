package dates

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{"epoch", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), "1970-01-01 00:00:00"},
		{"sample", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC), "2026-05-21 14:30:00"},
		{"end-of-year", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), "2026-12-31 23:59:59"},
		{"single-digit-padding", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), "2026-01-02 03:04:05"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDate(tc.in)
			if got != tc.want {
				t.Errorf("FormatDate(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDateValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Time
	}{
		{"epoch", "1970-01-01 00:00:00", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"sample", "2026-05-21 14:30:00", time.Date(2026, 5, 21, 14, 30, 0, 0, time.UTC)},
		{"end-of-year", "2026-12-31 23:59:59", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDate(tc.in)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", tc.in, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("ParseDate(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDateInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"not-a-date", "not-a-date"},
		{"wrong-shape", "2026/05/21 14:30:00"},
		{"out-of-range", "2026-13-01 12:00:00"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDate(tc.in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
			if !strings.Contains(err.Error(), tc.in) && tc.in != "" {
				t.Errorf("error message %q doesn't mention input %q", err.Error(), tc.in)
			}
			if errors.Unwrap(err) == nil {
				t.Errorf("expected wrapped error, got %v", err)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := time.Date(2026, 5, 21, 14, 30, 45, 0, time.UTC)
	s := FormatDate(original)
	parsed, err := ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	if !parsed.Equal(original) {
		t.Errorf("round-trip: parsed=%v, original=%v", parsed, original)
	}
}
