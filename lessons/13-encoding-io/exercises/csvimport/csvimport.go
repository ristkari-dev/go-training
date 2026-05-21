// Package csvimport reads expenses from a simple CSV format.
//
// Format: one expense per line, three comma-separated fields:
//
//	date,amount,category
//	2026-05-21,4.50,coffee
//	2026-05-21,12.00,lunch
//
// No header row. No quoting. No embedded commas. Blank lines are
// skipped. Leading/trailing whitespace on each field is trimmed.
//
// Why this format: it's the simplest thing that demonstrates the
// pattern (read lines → parse → assemble structs). Real CSV with
// quoting and escaping is handled by the encoding/csv package — out of
// scope for this lesson, which focuses on io.Reader + bufio.Scanner.
package csvimport

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/expense"
)

// Parse reads CSV from r and returns the parsed expenses.
//
// Returns the partial list parsed so far PLUS a non-nil error if any
// row is malformed. The error message includes the line number.
//
// Examples:
//
//	Parse(strings.NewReader(""))                              → ([]Expense{}, nil)
//	Parse(strings.NewReader("2026-05-21,4.50,coffee"))        → (one expense, nil)
//	Parse(strings.NewReader("...\n2026-05-21,oops,coffee"))   → (partial, wrapped strconv error mentioning line 2)
//
// Hint:
//  1. s := bufio.NewScanner(r)
//  2. var out []expense.Expense
//  3. for line := 1; s.Scan(); line++ {
//     text := strings.TrimSpace(s.Text())
//     if text == "" { continue }   // skip blank lines
//     e, err := parseLine(text)
//     if err != nil → return out, fmt.Errorf("csvimport: line %d: %w", line, err)
//     out = append(out, e)
//     }
//  4. if err := s.Err(); err != nil → return out, fmt.Errorf("csvimport: %w", err)
//  5. return out, nil
//
// parseLine is a helper: split on comma, expect 3 fields, ParseFloat
// the amount.
func Parse(r io.Reader) ([]expense.Expense, error) {
	_ = bufio.NewScanner
	_ = strings.TrimSpace
	_ = strconv.ParseFloat
	_ = fmt.Errorf
	panic("TODO: scan loop with TrimSpace + skip-blank + parseLine + line-numbered wrapping")
}
