// Package exercises is the starter code for lesson 06: Structs and methods.
//
// This file holds the MAIN exercise. Define the Expense struct, give it
// two methods (Format and IsHigh), and write the TotalsByCategory function
// that aggregates a []Expense by category.
package exercises

// Expense is one row of the expense tracker — a single charge with a date,
// amount, and category.
//
// All three fields are exported (capitalised) because callers in the test
// file need to construct Expense values directly using struct literals.
// Lesson 07 covers exported vs unexported formally.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same format as lesson 04's FormatExpense — now bound to
// the Expense receiver.
//
// Examples:
//
//	Expense{Date: "2026-05-12", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-12  €4.50    coffee"
//	Expense{Date: "2026-05-12", Amount: 12, Category: "lunch"}.Format()
//	  → "2026-05-12  €12.00   lunch"
//	Expense{Date: "2026-05-12", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-12  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Examples:
//
//	Expense{Amount: 4.50}.IsHigh()  → false
//	Expense{Amount: 50}.IsHigh()    → false   (boundary: 50 is NOT high)
//	Expense{Amount: 50.01}.IsHigh() → true
//	Expense{Amount: 999}.IsHigh()   → true
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// TotalsByCategory returns a map of category → summed amount, built from es.
//
// Empty or nil input returns an empty (non-nil) map — not an error. Callers
// can range over it normally.
//
// Examples:
//
//	es := []Expense{
//		{Date: "2026-05-12", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-12", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-12", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]   (non-nil empty map)
//
// Hint: initialise `totals := map[string]float64{}`, then range over es
// accumulating totals[e.Category] += e.Amount. Lesson 05's idiom works here.
func TotalsByCategory(es []Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into a map[string]float64")
}
