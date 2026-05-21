package store

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/13-encoding-io/exercises/expense"
)

// runStoreContract runs the same contract tests against both JSONStore
// and MemoryStore. Skeleton tests reference all the helpers but don't
// actually call the panicking JSONStore methods.
func runStoreContract(t *testing.T, name string, build func() Store) {
	t.Run(name+"/Save_then_Load_roundtrips", func(t *testing.T) {
		// TODO: s := build(); save sample expenses; load them back; assert equal.
		_ = build
		_ = reflect.DeepEqual
	})
	t.Run(name+"/Save_replaces_existing", func(t *testing.T) {
		// TODO: save a set, then save a different set, load, assert only the second set is present.
		_ = build
	})
}

func TestJSONStore(t *testing.T) {
	// TODO: tmp dir + runStoreContract with a JSONStore at tmp/expenses.json
	_ = filepath.Join
	_ = runStoreContract
}

func TestMemoryStore(t *testing.T) {
	// TODO: runStoreContract(t, "MemoryStore", func() Store { return NewMemoryStore() })
	_ = expense.Expense{}
}

// TestJSONStoreLoadReturnsErrNotFound: missing file → errors.Is(err, ErrNotFound).
func TestJSONStoreLoadReturnsErrNotFound(t *testing.T) {
	// TODO:
	//   s := NewJSONStore(filepath.Join(t.TempDir(), "missing.json"))
	//   _, err := s.Load()
	//   if !errors.Is(err, ErrNotFound) → t.Errorf("...")
	_ = errors.Is
}

// TestJSONStoreLoadReturnsParseError: malformed JSON → errors.As to *ParseError.
func TestJSONStoreLoadReturnsParseError(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "bad.json")
	//   os.WriteFile(path, []byte("not-json"), 0o644)
	//   s := NewJSONStore(path)
	//   _, err := s.Load()
	//   var pe *ParseError; if !errors.As(err, &pe) → t.Errorf("...")
	//   if pe.Path != path → t.Errorf("...")
	//   if pe.Line < 1 → t.Errorf("expected line >= 1, got %d", pe.Line)
}

// TestMemoryStoreLoadNeverReturnsErrNotFound: empty MemoryStore is valid state.
func TestMemoryStoreLoadNeverReturnsErrNotFound(t *testing.T) {
	// TODO:
	//   s := NewMemoryStore()
	//   _, err := s.Load()
	//   if errors.Is(err, ErrNotFound) → t.Error("MemoryStore should never return ErrNotFound")
}
