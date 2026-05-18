// Package clock provides a tiny time helper for the lesson 08 warm-up.
//
// This is the simplest possible "publish a function from a subpackage"
// demo. The matching cmd/timestamp/main.go imports clock and prints
// the result of Now().
package clock

// Now returns the current UTC time formatted as "2006-01-02 15:04:05".
//
// Examples:
//
//	clock.Now()  → "2026-05-18 14:30:00"   (whatever the current UTC time is)
//
// Hint: import "time", then time.Now().UTC().Format("2006-01-02 15:04:05").
// The string "2006-01-02 15:04:05" is Go's reference time — a quirky way
// to specify date/time formats. See pkg.go.dev/time#Time.Format if curious.
func Now() string {
	panic("TODO: return time.Now().UTC().Format(\"2006-01-02 15:04:05\")")
}
