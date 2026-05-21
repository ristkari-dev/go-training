// Package store is the lesson 11 reference implementation.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/11-errors/solutions/expense"
)

// ErrNotFound is returned (wrapped) by JSONStore.Load when the backing
// file doesn't exist yet.
var ErrNotFound = errors.New("store: not found")

// ParseError wraps a JSON unmarshal failure with file path + line.
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("store parse %s:%d: %v", e.Path, e.Line, e.Cause)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

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

// Load reads the JSON file at j.Path. Returns wrapped ErrNotFound if the
// file doesn't exist; *ParseError if the JSON is malformed.
func (j *JSONStore) Load() ([]expense.Expense, error) {
	data, err := os.ReadFile(j.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", j.Path, err)
	}
	var es []expense.Expense
	if err := json.Unmarshal(data, &es); err != nil {
		return nil, &ParseError{Path: j.Path, Line: lineFromJSONErr(data, err), Cause: err}
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

func lineFromJSONErr(data []byte, err error) int {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return lineFromOffset(data, syntaxErr.Offset)
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return lineFromOffset(data, typeErr.Offset)
	}
	return 0
}

func lineFromOffset(data []byte, offset int64) int {
	line := 1
	for i := int64(0); i < offset && int(i) < len(data); i++ {
		if data[i] == '\n' {
			line++
		}
	}
	return line
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
