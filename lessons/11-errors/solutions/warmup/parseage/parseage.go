// Package parseage is the lesson 11 warm-up reference implementation.
package parseage

import (
	"fmt"
	"strconv"
)

// ParseAge parses s as an integer age. Wraps strconv.Atoi errors with the
// offending input string.
func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parseage: invalid age %q: %w", s, err)
	}
	return n, nil
}
