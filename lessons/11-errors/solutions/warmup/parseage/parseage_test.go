package parseage

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestParseAgeValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"positive", "25", 25},
		{"zero", "0", 0},
		{"negative", "-1", -1},
		{"large", "1000000", 1000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAge(tc.in)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseAge(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseAgeInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"non-numeric", "twenty"},
		{"trailing-space", "25 "},
		{"with-letters", "25a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseAge(tc.in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
			if !strings.Contains(err.Error(), tc.in) {
				t.Errorf("error message %q doesn't mention input %q", err.Error(), tc.in)
			}
			if !errors.Is(err, strconv.ErrSyntax) {
				t.Errorf("err = %v; expected errors.Is(err, strconv.ErrSyntax) to be true", err)
			}
		})
	}
}
