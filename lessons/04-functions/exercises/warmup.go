// Package exercises is the starter code for lesson 04: Functions & first tests.
//
// This file holds the WARM-UP exercise. Two small functions to build muscle
// memory for "function with multiple returns" and "variadic argument list,
// with an error return for empty input." Tests are pre-written — fix the
// implementations until they pass.
package exercises

import "errors"

// errEmptyMinMax is returned by MinMax when called with no values.
//
// Sentinel-style error variables are a common Go pattern: declare once with
// errors.New, return as-is. We'll see more sophisticated error styles in
// lesson 11.
var errEmptyMinMax = errors.New("MinMax: requires at least one value")

// Add returns the sum of two ints.
//
// Examples:
//
//	Add(2, 3)   → 5
//	Add(-1, 1)  → 0
func Add(a, b int) int {
	panic("TODO: return a + b")
}

// MinMax returns the smallest and largest of the given ints, plus an error.
//
// When called with at least one int, it returns (min, max, nil). When called
// with zero ints, it returns (0, 0, errEmptyMinMax).
//
// Examples:
//
//	MinMax(3, 1, 4, 1, 5, 9, 2, 6)  → (1, 9, nil)
//	MinMax(42)                       → (42, 42, nil)
//	MinMax()                         → (0, 0, error)
//
// Hint: variadic parameters arrive as a slice — `xs ...int` is `xs []int`
// inside the function body. Walk the slice with `for i, v := range xs` or
// `for _, v := range xs`. The first element is your initial min and max;
// fold the rest in.
func MinMax(xs ...int) (int, int, error) {
	panic("TODO: return min, max for non-empty input; return (0, 0, errEmptyMinMax) for empty input")
}
