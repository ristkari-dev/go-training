// Package solutions is the reference implementation for lesson 05: Slices and maps.
package solutions

import "errors"

// errEmptyInputs is returned by TotalsByCategory when either input slice
// is empty.
var errEmptyInputs = errors.New("TotalsByCategory: amounts and categories must be non-empty")

// errLengthMismatch is returned by TotalsByCategory when the two input
// slices have different lengths.
var errLengthMismatch = errors.New("TotalsByCategory: amounts and categories must have the same length")

// TotalsByCategory aggregates parallel amount/category slices into a per-
// category total map. See the doc comment in exercises/main.go for details.
func TotalsByCategory(amounts []float64, categories []string) (map[string]float64, error) {
	if len(amounts) == 0 || len(categories) == 0 {
		return nil, errEmptyInputs
	}
	if len(amounts) != len(categories) {
		return nil, errLengthMismatch
	}
	totals := map[string]float64{}
	for i, cat := range categories {
		totals[cat] += amounts[i]
	}
	return totals, nil
}
