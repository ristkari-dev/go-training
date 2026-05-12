// Package solutions is the reference implementation for lesson 04: Functions & first tests.
//
// This file holds the warm-up reference solution.
package solutions

import "errors"

// errEmptyMinMax is returned by MinMax when called with no values.
var errEmptyMinMax = errors.New("MinMax: requires at least one value")

// Add returns the sum of two ints.
func Add(a, b int) int {
	return a + b
}

// MinMax returns the smallest and largest of the given ints, plus an error
// for the empty-input case.
func MinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmptyMinMax
	}
	min, max := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, nil
}
