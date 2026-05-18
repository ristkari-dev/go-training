// Package summary provides aggregation and visualisation helpers for
// the lesson 08 capstone CLI. Three pure functions, no I/O.
package summary

import (
	"strings"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// TotalsByCategory returns a map of category → summed amount, built from es.
// Empty/nil input returns a non-nil empty map.
//
// Same behaviour as lesson 06/07's TotalsByCategory — repeated here as
// it's the natural fit for the capstone's "summary" command.
//
// Examples:
//
//	es := []expense.Expense{
//		{Date: "2026-05-18", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-18", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-18", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]
//
// Hint: same as before — initialise totals := map[string]float64{}, range
// over es, accumulate.
func TotalsByCategory(es []expense.Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into a map[string]float64")
}

// BiggestCategory returns the category with the largest total.
//
// If totals is empty or nil, BiggestCategory returns ("", 0) — there's no
// "biggest" of nothing.
//
// If two categories tie for biggest, BiggestCategory returns the one
// whose name sorts alphabetically first (so the output is deterministic
// regardless of map iteration order).
//
// Examples:
//
//	BiggestCategory(map[string]float64{"coffee": 14.49, "rent": 75})
//	  → ("rent", 75)
//	BiggestCategory(map[string]float64{"a": 10, "b": 10})
//	  → ("a", 10)  (alphabetical tie-break)
//	BiggestCategory(map[string]float64{})    → ("", 0)
//	BiggestCategory(nil)                      → ("", 0)
//
// Hint: range over the map keys, sort them to make iteration deterministic,
// then walk sorted keys tracking the running maximum.
func BiggestCategory(totals map[string]float64) (string, float64) {
	panic("TODO: return the category with the largest total; alphabetical tie-break; (\"\", 0) for empty")
}

// Bar returns a string of up to `width` Unicode block characters
// proportional to value/max. Uses '█' (U+2588 FULL BLOCK) for full
// units and '▌' (U+258C LEFT HALF BLOCK) for half units.
//
// The result has between 0 and `width` chars. A half-block appears when
// the fractional remainder is >= 0.5 of a unit.
//
// Returns "" when:
//   - max <= 0 (can't compute a ratio)
//   - value <= 0 (no bar to draw)
//   - width <= 0 (no room)
//
// Examples (width=8):
//
//	Bar(0, 10, 8)       → ""
//	Bar(10, 10, 8)      → "████████"        (full)
//	Bar(5, 10, 8)       → "████"            (half → 4 full chars)
//	Bar(2.5, 10, 8)     → "██"              (quarter → 2 full)
//	Bar(0.5, 8, 8)      → "▌"               (just half)
//	Bar(1, 8, 8)        → "█"               (1 full)
//
// Hint: units := value / max * float64(width); full := int(units);
// half := units - float64(full) >= 0.5;
// result := strings.Repeat("█", full); if half { result += "▌" }.
func Bar(value, max float64, width int) string {
	_ = strings.Repeat // keep the import compiling until you use it
	panic("TODO: build the bar string per the doc comment")
}
