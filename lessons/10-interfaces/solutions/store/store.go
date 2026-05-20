// Package store is the lesson 10 reference implementation.
package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/solutions/expense"
)

// Store abstracts over expense persistence.
type Store interface {
	Load() ([]expense.Expense, error)
	Save(es []expense.Expense) error
}

// JSONStore persists expenses as pretty-printed JSON at Path.
type JSONStore struct {
	Path string
}

// NewJSONStore constructs a JSONStore at the given path.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

// Load reads the JSON file at j.Path. Returns ([], nil) if the file
// doesn't exist.
func (j *JSONStore) Load() ([]expense.Expense, error) {
	data, err := os.ReadFile(j.Path)
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

// Save writes es to j.Path as pretty-printed JSON.
func (j *JSONStore) Save(es []expense.Expense) error {
	data, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(j.Path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(j.Path, data, 0o644)
}

// MemoryStore keeps expenses in memory.
type MemoryStore struct {
	items []expense.Expense
}

// NewMemoryStore constructs an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: []expense.Expense{}}
}

// Load returns a defensive copy of the stored expenses.
func (m *MemoryStore) Load() ([]expense.Expense, error) {
	return append([]expense.Expense{}, m.items...), nil
}

// Save replaces the stored expenses with a copy of es.
func (m *MemoryStore) Save(es []expense.Expense) error {
	m.items = append([]expense.Expense{}, es...)
	return nil
}
