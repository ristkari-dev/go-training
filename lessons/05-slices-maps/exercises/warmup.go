// Package exercises is the starter code for lesson 05: Slices and maps.
//
// This file holds the WARM-UP exercise. Three functions over slices that
// also let you write your first table tests of your own (warmup_test.go
// ships as a skeleton — fill in the cases and assertion bodies).
package exercises

import "errors"

// errEmptyWarmupMax is returned by WarmupMax when called with no values.
var errEmptyWarmupMax = errors.New("WarmupMax: requires at least one value")

// WarmupSum returns the sum of the integers in xs.
//
// An empty slice returns 0 — the additive identity is a well-defined answer.
//
// Examples:
//
//	WarmupSum([]int{1, 2, 3})  → 6
//	WarmupSum([]int{})         → 0
//	WarmupSum(nil)             → 0   (nil slice walks zero times under range)
//
// Hint: `for _, v := range xs { total += v }` is the canonical shape.
func WarmupSum(xs []int) int {
	panic("TODO: walk xs with range, accumulate into a total, return it")
}

// WarmupMax returns the largest int in xs, plus an error.
//
// Non-empty input returns (max, nil). Empty input returns (0, errEmptyWarmupMax).
// Unlike WarmupSum, "max of nothing" has no defensible default — hence the error.
//
// Examples:
//
//	WarmupMax([]int{3, 1, 4, 1, 5, 9, 2, 6})  → (9, nil)
//	WarmupMax([]int{-3, -1, -7})              → (-1, nil)
//	WarmupMax([]int{42})                      → (42, nil)
//	WarmupMax([]int{})                        → (0, error)
//
// Hint: guard the empty case first (early return). Then seed max with
// xs[0] and walk xs[1:] with range, updating when a larger value is found.
func WarmupMax(xs []int) (int, error) {
	panic("TODO: return error for empty; otherwise return the maximum")
}

// WarmupUnique returns the values of xs in their original order with
// duplicates removed. The original slice is not modified.
//
// Examples:
//
//	WarmupUnique([]string{"a", "b", "a", "c", "b"})  → ["a", "b", "c"]
//	WarmupUnique([]string{"x"})                      → ["x"]
//	WarmupUnique([]string{})                         → []   (non-nil empty slice)
//	WarmupUnique(nil)                                → []
//
// Hint: a `map[string]bool` makes a great "seen" set. For each value,
// check `seen[v]`; if not seen, append to the result and set `seen[v] = true`.
// Always return a non-nil slice — initialise with `result := []string{}`.
func WarmupUnique(xs []string) []string {
	panic("TODO: dedupe while preserving order; use a map[string]bool as a seen-set")
}
