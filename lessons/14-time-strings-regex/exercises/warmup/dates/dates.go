// Package dates is the lesson 14 warm-up: format and parse dates using
// the famous Go reference-time idiom.
//
// Go's date formatting uses the literal reference time
//
//	Mon Jan 2 15:04:05 MST 2006
//
// (which numerically is 01/02 03:04:05PM '06 MST). Instead of cryptic
// %Y-%m-%d directives, you write what the OUTPUT should look like in
// terms of that reference moment. To format "year-month-day" you write
// "2006-01-02"; to format "hour:minute:second" you write "15:04:05".
//
// This package teaches the round trip — format a time, parse it back,
// get the same time (truncated to seconds).
package dates

import (
	"strconv"
	"time"
)

// FormatDate returns t formatted as "YYYY-MM-DD HH:MM:SS" (e.g.,
// "2026-05-21 14:30:00"). Uses the reference-time idiom: the layout
// string "2006-01-02 15:04:05" describes the OUTPUT shape using the
// reference moment's components.
//
// Hint: return t.Format("2006-01-02 15:04:05")
func FormatDate(t time.Time) string {
	_ = time.Now
	_ = strconv.Itoa
	panic("TODO: t.Format with the reference time layout")
}

// ParseDate parses s as a "YYYY-MM-DD HH:MM:SS" date.
//
// On success returns (parsed time, nil). On failure returns
// (zero time, wrapped error mentioning the offending input).
//
// Hint:
//  1. t, err := time.Parse("2006-01-02 15:04:05", s)
//  2. if err != nil → return time.Time{}, fmt.Errorf("dates: invalid date %q: %w", s, err)
//  3. return t, nil
//
// time.Parse uses the SAME reference-time layout string as Format. The
// trick: same string for both directions.
func ParseDate(s string) (time.Time, error) {
	panic("TODO: time.Parse with the reference time layout; wrap errors with %w")
}
