// Package clock is the lesson 08 warm-up reference implementation.
package clock

import "time"

// Now returns the current UTC time formatted as "2006-01-02 15:04:05".
func Now() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}
