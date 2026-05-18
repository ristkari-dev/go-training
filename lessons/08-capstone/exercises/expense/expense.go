// Package expense holds the Expense type + methods for lesson 08's capstone.
//
// This is the same code you wrote in lesson 07 — you may copy from
// lessons/07-packages/exercises/expense/ or re-implement here. Lesson 08's
// summary, storage, and cmd/expenses packages all import expense.Expense.
package expense

// Expense is one row of the expense tracker — date, amount, category.
//
// All three fields are exported because the test file, the summary and
// storage subpackages, and the CLI binary all construct Expense values
// with struct literals or read them from JSON.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" using fmt.Sprintf with
// "%s  €%-7.2f %s". Same byte-for-byte output as lesson 04/06/07's Format.
//
// Examples:
//
//	Expense{Date: "2026-05-18", Amount: 4.50, Category: "coffee"}.Format()
//	  → "2026-05-18  €4.50    coffee"
//	Expense{Date: "2026-05-18", Amount: 999.99, Category: "rent"}.Format()
//	  → "2026-05-18  €999.99  rent"
//
// Hint: fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category).
func (e Expense) Format() string {
	panic("TODO: return fmt.Sprintf-formatted string per the doc comment")
}

// IsHigh reports whether the expense is over €50.
//
// Boundary: amount of exactly 50 is NOT high; 50.01 IS high.
func (e Expense) IsHigh() bool {
	panic("TODO: return e.Amount > 50")
}
