// Package store is the lesson 15 reference implementation.
// JSONStore uses streaming Encoder/Decoder (refactored from L11's
// Marshal/Unmarshal). Observable behavior preserved.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

var ErrNotFound = errors.New("store: not found")

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

type Store interface {
	Load() ([]expense.Expense, error)
	Save(es []expense.Expense) error
}

type JSONStore struct {
	Path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

func (j *JSONStore) Load() ([]expense.Expense, error) {
	f, err := os.Open(j.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", j.Path, err)
	}
	defer f.Close()

	var es []expense.Expense
	dec := json.NewDecoder(f)
	if err := dec.Decode(&es); err != nil {
		return nil, &ParseError{Path: j.Path, Line: lineFromDecoder(dec, err), Cause: err}
	}
	if es == nil {
		es = []expense.Expense{}
	}
	return es, nil
}

func (j *JSONStore) Save(es []expense.Expense) error {
	if dir := filepath.Dir(j.Path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(j.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(es)
}

func lineFromDecoder(dec *json.Decoder, err error) int {
	offset := dec.InputOffset()
	if offset <= 0 {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			offset = syntaxErr.Offset
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			offset = typeErr.Offset
		}
	}
	if offset <= 0 {
		return 1
	}
	return 1
}

type MemoryStore struct {
	items []expense.Expense
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: []expense.Expense{}}
}

func (m *MemoryStore) Load() ([]expense.Expense, error) {
	return append([]expense.Expense{}, m.items...), nil
}

func (m *MemoryStore) Save(es []expense.Expense) error {
	m.items = append([]expense.Expense{}, es...)
	return nil
}
