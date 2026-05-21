// Package store defines the Store interface and two implementations
// (JSONStore + MemoryStore), plus rich error types for the lesson 11
// error-handling upgrade.
//
// Two new exports vs lesson 10:
//   - ErrNotFound — sentinel for "JSONStore couldn't find its file."
//     Wrapped by JSONStore.Load() with the file path for context.
//   - ParseError  — custom error type wrapping JSON unmarshal failures
//     with the path + line where the parse failed.
//
// Callers use errors.Is(err, store.ErrNotFound) to handle first-time use
// gracefully, and errors.As(err, &pe) to extract a *ParseError for
// detailed error messages.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/11-errors/exercises/expense"
)

// ErrNotFound is returned (wrapped) by JSONStore.Load when the backing
// file doesn't exist yet — typically because this is the first time the
// CLI has been run with this -file path. Callers use errors.Is to detect.
var ErrNotFound = errors.New("store: not found")

// ParseError wraps a JSON unmarshal failure with the file path and
// (approximate) line number where the parse failed.
//
// Fields are exported so callers can inspect them after extracting via
// errors.As. Unwrap returns Cause so errors.Is can chase through.
type ParseError struct {
	Path  string
	Line  int
	Cause error
}

// Error formats the error as "store parse <path>:<line>: <cause>".
//
// Hint: fmt.Sprintf with %s, %d, %v.
func (e *ParseError) Error() string {
	panic("TODO: return fmt.Sprintf with Path, Line, Cause")
}

// Unwrap returns Cause so errors.Is / errors.As can chase through.
//
// Hint: just return e.Cause.
func (e *ParseError) Unwrap() error {
	panic("TODO: return e.Cause")
}

// Store abstracts over expense persistence (unchanged from lesson 10).
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

// Load reads the JSON file at j.Path.
//
// CHANGED FROM LESSON 10: missing file now returns a wrapped ErrNotFound
// (caller can detect via errors.Is). Malformed JSON returns a *ParseError
// with path + line.
//
// Hint:
//  1. data, err := os.ReadFile(j.Path)
//  2. if errors.Is(err, fs.ErrNotExist) → return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
//  3. if err != nil → return nil, fmt.Errorf("loading %s: %w", j.Path, err)
//  4. var es []expense.Expense
//  5. if err := json.Unmarshal(data, &es); err != nil → return nil,
//     &ParseError{Path: j.Path, Line: lineFromJSONErr(data, err), Cause: err}
//  6. if es == nil → es = []expense.Expense{}
//  7. return es, nil
func (j *JSONStore) Load() ([]expense.Expense, error) {
	panic("TODO: see hint in the doc comment")
}

// Save writes es to j.Path as pretty-printed JSON (unchanged from L10).
//
// Hint: json.MarshalIndent; os.MkdirAll the parent dir; os.WriteFile.
func (j *JSONStore) Save(es []expense.Expense) error {
	panic("TODO: marshal with indent; create parent dir; write file")
}

// lineFromJSONErr returns the (1-based) line number where err occurred,
// based on the byte offset in *json.SyntaxError or *json.UnmarshalTypeError.
// Returns 0 if the error doesn't carry an offset.
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

// lineFromOffset counts newlines in data[0..offset] and returns 1 +
// that count (1-based line numbering).
func lineFromOffset(data []byte, offset int64) int {
	line := 1
	for i := int64(0); i < offset && int(i) < len(data); i++ {
		if data[i] == '\n' {
			line++
		}
	}
	return line
}

// MemoryStore keeps expenses in memory. Unchanged from L10 — does NOT
// return ErrNotFound (an empty MemoryStore is a valid empty state, not
// an error). The contract test demonstrates this divergence.
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

// _ keeps imports used while function bodies are TODO. Students remove
// this when their implementations actually use these.
var (
	_ = filepath.Join
	_ = fs.ErrNotExist
	_ = json.Unmarshal
	_ = os.ReadFile
	_ = errors.Is
	_ = fmt.Errorf
)
