// Package expense holds the Expense type + methods for lesson 07's main exercise.
//
// In lesson 06 these all lived at the package level alongside the test
// scaffolding. Here we extract them into their own subpackage so the
// top-level main.go imports `expense` and uses the type via its package
// qualifier (expense.Expense, expense.TotalsByCategory). That's how
// real-world Go code organises types you want to reuse across binaries.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// All three fields are exported (capitalised) because the parent package's
// main.go and the test file in this package construct Expense values with
// struct literals. If a field were lower-case, neither could set it.
type Expense struct {
	Date     string
	Amount   float64
	Category string
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04's FormatExpense
// and lesson 06's Expense.Format — now in its own package.
//
// Examples:
//
//	Expense{Date: "2026-05-15", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-15  €4.50    coffee"
//	Expense{Date: "2026-05-15", Amount: 12, Category: "lunch"}.Format()
//	  → "2026-05-15  €12.00   lunch"
//	Expense{Date: "2026-05-15", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-15  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Examples:
//
//	Expense{Amount: 50}.IsHigh()    → false   (boundary: 50 is NOT high)
//	Expense{Amount: 50.01}.IsHigh() → true
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}

// TotalsByCategory returns a map of category → summed amount, built from es.
// Empty/nil input returns a non-nil empty map.
//
// Examples:
//
//	es := []Expense{
//		{Date: "2026-05-15", Amount: 4.50, Category: "coffee"},
//		{Date: "2026-05-15", Amount: 12, Category: "lunch"},
//		{Date: "2026-05-15", Amount: 9.99, Category: "coffee"},
//	}
//	TotalsByCategory(es) → map[coffee:14.49 lunch:12]
//	TotalsByCategory(nil) → map[]   (non-nil empty map)
//
// Hint: same idiom as lesson 06's TotalsByCategory — initialise totals
// as map[string]float64{}, range over es, accumulate.
func TotalsByCategory(es []Expense) map[string]float64 {
	panic("TODO: walk es with range, accumulate into map[string]float64")
}
