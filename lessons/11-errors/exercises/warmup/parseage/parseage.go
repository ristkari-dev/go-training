// Package parseage is the lesson 11 warm-up: a tiny demo of error wrapping.
//
// ParseAge takes a string and returns an int. When the input isn't a valid
// integer, ParseAge returns an error that:
//   - Mentions the offending input in its message (for humans).
//   - Wraps the underlying strconv error (so errors.Is(err, strconv.ErrSyntax)
//     still works for callers that want to discriminate).
//
// Both pieces matter. A bare strconv.Atoi error tells you "syntax error" but
// not WHICH input was bad. A new errors.New("invalid age") loses the
// underlying sentinel that callers might check.
package parseage

import "strconv"

// ParseAge parses s as a non-negative integer age.
//
// On success, returns (age, nil).
//
// On failure, returns (0, wrapped error). The wrapped error mentions the
// offending input AND preserves the underlying strconv error via %w so
// errors.Is(err, strconv.ErrSyntax) returns true.
//
// Examples:
//
//	ParseAge("25")      → (25, nil)
//	ParseAge("")        → (0, error)  // errors.Is(err, strconv.ErrSyntax) == true
//	ParseAge("twenty")  → (0, error)  // same
//	ParseAge("-1")      → (-1, nil)   // we don't validate non-negative; the underlying Atoi succeeds
//
// Hint: strconv.Atoi(s); if err != nil return fmt.Errorf("parseage: invalid
// age %q: %w", s, err).
func ParseAge(s string) (int, error) {
	_ = strconv.Atoi
	panic("TODO: call strconv.Atoi; on error return fmt.Errorf with %q and %w wrapping")
}
