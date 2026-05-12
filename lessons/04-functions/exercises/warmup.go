// Package exercises is the starter code for lesson 04: Functions & first tests.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "function with multiple returns" and "variadic argument list,
// with an error return for empty input." Tests are pre-written — fix the
// implementations until they pass.
package exercises

import "errors"

// errEmptyWarmupMinMax is returned by WarmupMinMax when called with no values.
//
// Sentinel-style error variables are a common Go pattern: declare once with
// errors.New, return as-is. We'll see more sophisticated error styles in
// lesson 11.
var errEmptyWarmupMinMax = errors.New("WarmupMinMax: requires at least one value")

// WarmupAdd returns the sum of two ints.
//
// Examples:
//
//	WarmupAdd(2, 3)   → 5
//	WarmupAdd(-1, 1)  → 0
func WarmupAdd(a, b int) int {
	panic("TODO: return a + b")
}

// WarmupMinMax returns the smallest and largest of the given ints, plus an
// error.
//
// When called with at least one int, it returns (min, max, nil). When called
// with zero ints, it returns (0, 0, errEmptyWarmupMinMax).
//
// Examples:
//
//	WarmupMinMax(3, 1, 4, 1, 5, 9, 2, 6)  → (1, 9, nil)
//	WarmupMinMax(42)                      → (42, 42, nil)
//	WarmupMinMax()                        → (0, 0, error)
//
// Hint: variadic parameters arrive as a slice — `xs ...int` is `xs []int`
// inside the function body. Walk the slice with `for i, v := range xs` or
// `for _, v := range xs`. The first element is your initial min and max;
// fold the rest in.
func WarmupMinMax(xs ...int) (int, int, error) {
	panic("TODO: return min, max for non-empty input; return (0, 0, errEmptyWarmupMinMax) for empty input")
}
