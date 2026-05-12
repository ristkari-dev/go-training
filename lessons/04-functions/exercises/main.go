// Package exercises is the starter code for lesson 04: Functions & first tests.
//
// This file holds the MAIN exercise. Two functions to implement, AND the
// matching tests for them in main_test.go. The test file ships with a
// skeleton table test for each function — you fill in the cases and the
// loop body. See exercises/main_test.go and the README's "Exercise: main"
// section for the test-writing walk-through.
package exercises

// Categorise returns "snack" for amounts under €10, "regular" for €10
// (inclusive) up to €50 (inclusive), and "splurge" above €50.
//
// Same behaviour as lesson 03's Categorise — the focus here is on writing
// the *test* for it, in table form. Use whichever shape (if-chain or
// tagless switch) you prefer.
//
// Examples:
//
//	Categorise(4.50)  → "snack"
//	Categorise(50.00) → "regular"   (boundary: 50 is the upper edge of regular)
//	Categorise(75.00) → "splurge"
func Categorise(amount float64) string {
	panic("TODO: return \"snack\" / \"regular\" / \"splurge\" by amount")
}

// FormatExpense returns a single-line, column-aligned rendering of one
// expense.
//
// Format: "YYYY-MM-DD  €AMOUNT  category"
//   - the date is the caller's responsibility — it's printed as-is
//   - the amount is prefixed with €, has 2 decimal places, and is left-padded
//     to 7 chars total (the "%-7.2f" verb in fmt.Sprintf)
//   - the category follows after two spaces
//
// Examples:
//
//	FormatExpense("2026-05-12", 4.50, "coffee")   → "2026-05-12  €4.50    coffee"
//	FormatExpense("2026-05-12", 12.00, "lunch")   → "2026-05-12  €12.00   lunch"
//	FormatExpense("2026-05-12", 999.99, "rent")   → "2026-05-12  €999.99  rent"
//
// Hint: fmt.Sprintf with "%s  €%-7.2f %s" (note the two spaces, the leading
// €, the width specifier, and the single trailing space) produces the
// expected output for the examples above.
func FormatExpense(date string, amount float64, cat string) string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}
