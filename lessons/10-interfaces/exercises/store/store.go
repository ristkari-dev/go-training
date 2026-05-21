// Package store defines the Store interface for persisting expenses, plus
// two implementations: JSONStore (file-backed) and MemoryStore (in-process).
//
// The Store interface is the lesson's "aha": cmd/expenses doesn't care
// whether the data lives in a file or in memory — it just calls Load()
// and Save(). At runtime, a -store=mem|json flag picks the concrete impl.
package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/expense"
)

// Store abstracts over expense persistence. Implementations must satisfy
// both methods.
type Store interface {
	// Load returns all expenses. If there are no expenses (e.g. the file
	// doesn't exist for JSONStore, or Save has never been called for
	// MemoryStore), returns an empty slice and nil error.
	Load() ([]expense.Expense, error)

	// Save replaces the persisted expenses with es.
	Save(es []expense.Expense) error
}

// JSONStore persists expenses as pretty-printed JSON at Path.
//
// If Path doesn't exist when Load is called, Load returns ([], nil) —
// the same "first-time use is friendly" behaviour as lesson 08's storage.
type JSONStore struct {
	Path string
}

// NewJSONStore constructs a JSONStore at the given path.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{Path: path}
}

// Load reads the JSON file at j.Path and returns its contents.
//
// Hint: re-use the lesson 08 storage logic. Read the file with
// os.ReadFile; if errors.Is(err, fs.ErrNotExist), return an empty
// slice + nil error. Otherwise json.Unmarshal into a []expense.Expense
// and return it.
func (j *JSONStore) Load() ([]expense.Expense, error) {
	_ = json.Unmarshal
	_ = errors.Is
	_ = fs.ErrNotExist
	panic("TODO: read j.Path; on fs.ErrNotExist return ([], nil); else json.Unmarshal")
}

// Save writes es to j.Path as pretty-printed JSON.
//
// Hint: json.MarshalIndent(es, "", "  "); os.MkdirAll for the parent
// directory if it doesn't exist; os.WriteFile to write the bytes.
func (j *JSONStore) Save(es []expense.Expense) error {
	_ = filepath.Dir
	_ = os.MkdirAll
	panic("TODO: marshal es with indent; create parent dir; write file")
}

// MemoryStore keeps expenses in memory. Data does not persist across
// process restarts — useful mostly for tests where each test wants a
// fresh, fast Store with no on-disk state.
//
// Note: not concurrency-safe. Lesson 18 (sync) returns to this and adds
// a mutex.
type MemoryStore struct {
	items []expense.Expense
}

// NewMemoryStore constructs an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: []expense.Expense{}}
}

// Load returns a copy of the stored expenses. Returns ([]expense.Expense{},
// nil) when nothing has been saved yet.
//
// Hint: return a copy (not the live items slice) so callers can't mutate
// MemoryStore's internal state by appending to the returned slice. Use
// `append([]expense.Expense{}, m.items...)`.
func (m *MemoryStore) Load() ([]expense.Expense, error) {
	panic("TODO: return a copy of m.items + nil")
}

// Save replaces the stored expenses with a copy of es.
//
// Hint: same defensive-copy trick on the way in. `m.items = append(
// []expense.Expense{}, es...)`.
func (m *MemoryStore) Save(es []expense.Expense) error {
	panic("TODO: replace m.items with a copy of es; return nil")
}
