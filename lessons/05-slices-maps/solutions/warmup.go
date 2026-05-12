// Package solutions is the reference implementation for lesson 05: Slices and maps.
//
// This file holds the warm-up reference solution.
package solutions

import "errors"

// errEmptyWarmupMax is returned by WarmupMax when called with no values.
var errEmptyWarmupMax = errors.New("WarmupMax: requires at least one value")

// WarmupSum returns the sum of the integers in xs. Empty/nil input → 0.
func WarmupSum(xs []int) int {
	total := 0
	for _, v := range xs {
		total += v
	}
	return total
}

// WarmupMax returns the largest int in xs, plus an error for empty input.
func WarmupMax(xs []int) (int, error) {
	if len(xs) == 0 {
		return 0, errEmptyWarmupMax
	}
	max := xs[0]
	for _, v := range xs[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// WarmupUnique returns xs with duplicates removed, preserving original order.
// Always returns a non-nil slice.
func WarmupUnique(xs []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, v := range xs {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
