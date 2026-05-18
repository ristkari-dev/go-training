// Package storage handles JSON-backed persistence of Expense slices.
//
// This package is PROVIDED — you do not write or modify it for this lesson.
// We'll see how it works in Phase 2 (lesson 13 covers encoding/json
// properly). For now, treat it as a black box that loads and saves a
// []expense.Expense to a JSON file on disk.
//
// API:
//   - LoadExpenses(path) → ([]expense.Expense, error)
//     Returns an empty slice + nil error when the file doesn't exist
//     (so first-time use is friendly).
//   - SaveExpenses(path, es) → error
//     Writes pretty-printed JSON via json.MarshalIndent. Creates the
//     parent directory if needed.
package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/08-capstone/exercises/expense"
)

// LoadExpenses reads the JSON file at path and returns its contents.
//
// If the file doesn't exist, returns ([]expense.Expense{}, nil) — a clean
// empty slice, no error. This makes first-time use (no file yet) friendly:
// the caller can always range over the result.
//
// If the file exists but is unreadable or malformed, returns nil and an
// error wrapping the underlying cause.
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
//
// Creates the parent directory (with MkdirAll) if it doesn't exist.
// Overwrites the file if it does.
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
