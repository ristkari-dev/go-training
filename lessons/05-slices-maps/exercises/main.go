// Package exercises is the starter code for lesson 05: Slices and maps.
//
// This file holds the MAIN exercise. TotalsByCategory aggregates parallel
// expense slices into a per-category total map. The matching test file
// ships as a skeleton — fill in cases and assertions.
package exercises

import "errors"

// errEmptyInputs is returned by TotalsByCategory when either input slice
// is empty.
var errEmptyInputs = errors.New("TotalsByCategory: amounts and categories must be non-empty")

// errLengthMismatch is returned by TotalsByCategory when the two input
// slices have different lengths (they're meant to be parallel).
var errLengthMismatch = errors.New("TotalsByCategory: amounts and categories must have the same length")

// TotalsByCategory returns a map keyed by category whose values are the
// summed amounts for that category.
//
// Two error conditions, checked in order:
//  1. Empty inputs (either slice has length 0) → errEmptyInputs.
//  2. Length mismatch (len(amounts) != len(categories)) → errLengthMismatch.
//
// On success the map contains exactly the distinct categories from the input.
//
// Examples:
//
//	amounts    = []float64{4.50, 12, 75, 9.99}
//	categories = []string{"coffee", "lunch", "rent", "coffee"}
//	TotalsByCategory(amounts, categories) → map[coffee:14.49 lunch:12 rent:75], nil
//
//	TotalsByCategory(nil, nil)             → nil, errEmptyInputs
//	TotalsByCategory([]float64{1}, []string{"a", "b"}) → nil, errLengthMismatch
//
// Hint: guard the error cases first (early return). Then walk one of the
// slices with `for i, cat := range categories` (or amounts) and accumulate
// into a `map[string]float64`. Map zero-value is 0.0 for float64, so
// `totals[cat] += amounts[i]` works whether the key is present or not.
func TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error) {
	panic("TODO: validate inputs (empty + length mismatch); accumulate into map[string]float64")
}
