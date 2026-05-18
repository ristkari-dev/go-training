// Package storage handles JSON-backed persistence of Expense slices.
//
// (Solutions reference — identical to exercises/storage/storage.go except
// for the expense import path.)
package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/solutions/expense"
)

// LoadExpenses reads the JSON file at path and returns its contents.
// Returns ([]expense.Expense{}, nil) if the file doesn't exist.
func LoadExpenses(path string) ([]expense.Expense, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return []expense.Expense{}, nil
	}
	if err != nil {
		return nil, err
	}
	var es []expense.Expense
	if err := json.Unmarshal(data, &es); err != nil {
		return nil, err
	}
	if es == nil {
		es = []expense.Expense{}
	}
	return es, nil
}

// SaveExpenses writes es to path as pretty-printed JSON.
func SaveExpenses(path string, es []expense.Expense) error {
	data, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}
