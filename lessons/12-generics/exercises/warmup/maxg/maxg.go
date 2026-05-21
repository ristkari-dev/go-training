// Package maxg is the lesson 12 warm-up: a generic Max function.
//
// In lesson 04 (functions) you wrote separate Min/Max helpers for ints
// and floats — code duplication forced by Go's lack of generics back
// then. Go 1.18 added type parameters; Go 1.21 added the cmp.Ordered
// constraint to the standard library. Now we can write Max ONCE and use
// it on any ordered type (int, float64, string, time.Duration, ...).
//
// This is also a tiny demo of error handling for the empty-slice case
// (introduced in L11): an empty input returns the zero value of T plus
// a non-nil error.
package maxg

import "cmp"

// Max returns the largest element of xs.
//
// Type parameter T is constrained by cmp.Ordered — the set of types
// that support <, <=, >, >= (all integer types, all float types, strings).
// You CANNOT call Max on []bool or []struct{...} without an Ordered
// constraint — the compiler rejects it at the call site.
//
// On empty input, returns (zero T, non-nil error). On non-empty input,
// returns (largest element, nil).
//
// Type inference: callers usually write Max(xs) — the compiler infers T
// from the element type of xs. The explicit form Max[int](xs) is legal
// but unnecessary.
//
// Hint:
//  1. if len(xs) == 0 → return zero, errors.New("maxg: empty slice")
//     where `var zero T` gives the zero value for the inferred T.
//  2. m := xs[0]
//  3. for _, x := range xs[1:] → if x > m → m = x
//  4. return m, nil
func Max[T cmp.Ordered](xs []T) (T, error) {
	_ = cmp.Less[int]
	panic("TODO: handle empty, then linear scan; return largest element")
}
