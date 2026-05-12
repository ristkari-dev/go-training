// Package solutions is the reference implementation for lesson 04: Functions & first tests.
//
// This file holds the warm-up reference solution.
package solutions

import "errors"

// errEmptyWarmupMinMax is returned by WarmupMinMax when called with no values.
var errEmptyWarmupMinMax = errors.New("WarmupMinMax: requires at least one value")

// WarmupAdd returns the sum of two ints.
func WarmupAdd(a, b int) int {
	return a + b
}

// WarmupMinMax returns the smallest and largest of the given ints, plus an
// error for the empty-input case.
func WarmupMinMax(xs ...int) (int, int, error) {
	if len(xs) == 0 {
		return 0, 0, errEmptyWarmupMinMax
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
