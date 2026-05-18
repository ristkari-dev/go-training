// Package summary is the lesson 08 capstone reference implementation.
package summary

import (
	"sort"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
)

// TotalsByCategory returns a map of category → summed amount.
// Empty/nil input returns a non-nil empty map.
func TotalsByCategory(es []expense.Expense) map[string]float64 {
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}
	return totals
}

// BiggestCategory returns the category with the largest total. Alphabetical
// tie-break. ("", 0) for empty/nil input.
func BiggestCategory(totals map[string]float64) (string, float64) {
	if len(totals) == 0 {
		return "", 0
	}
	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	bestName := keys[0]
	bestTotal := totals[bestName]
	for _, k := range keys[1:] {
		if totals[k] > bestTotal {
			bestName = k
			bestTotal = totals[k]
		}
	}
	return bestName, bestTotal
}

// Bar returns a string of up to `width` Unicode block characters
// proportional to value/max.
func Bar(value, max float64, width int) string {
	if max <= 0 || value <= 0 || width <= 0 {
		return ""
	}
	units := value / max * float64(width)
	full := int(units)
	half := units-float64(full) >= 0.5
	result := strings.Repeat("█", full)
	if half {
		result += "▌"
	}
	return result
}
