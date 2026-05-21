// Package dates is the lesson 14 warm-up reference implementation.
package dates

import (
	"fmt"
	"time"
)

const layout = "2006-01-02 15:04:05"

// FormatDate returns t formatted with the lesson's "YYYY-MM-DD HH:MM:SS"
// layout, using Go's reference-time idiom.
func FormatDate(t time.Time) string {
	return t.Format(layout)
}

// ParseDate parses s with the lesson's layout. Wraps time.Parse errors
// with the offending input string.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("dates: invalid date %q: %w", s, err)
	}
	return t, nil
}
