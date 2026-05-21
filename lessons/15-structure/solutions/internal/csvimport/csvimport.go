// Package csvimport is the lesson 15 reference implementation.
package csvimport

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

// Parse reads CSV from r and returns the parsed expenses.
func Parse(r io.Reader) ([]expense.Expense, error) {
	s := bufio.NewScanner(r)
	out := []expense.Expense{}
	for line := 1; s.Scan(); line++ {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			continue
		}
		e, err := parseLine(text)
		if err != nil {
			return out, fmt.Errorf("csvimport: line %d: %w", line, err)
		}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return out, fmt.Errorf("csvimport: %w", err)
	}
	return out, nil
}

func parseLine(s string) (expense.Expense, error) {
	fields := strings.Split(s, ",")
	if len(fields) != 3 {
		return expense.Expense{}, fmt.Errorf("want 3 comma-separated fields, got %d", len(fields))
	}
	date := strings.TrimSpace(fields[0])
	amountStr := strings.TrimSpace(fields[1])
	category := strings.TrimSpace(fields[2])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("invalid amount %q: %w", amountStr, err)
	}
	return expense.Expense{Date: date, Amount: amount, Category: category}, nil
}
