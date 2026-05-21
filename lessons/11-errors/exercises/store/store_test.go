package store

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/11-errors/exercises/expense"
)

// runStoreContract verifies behaviours every Store implementation must
// support. Same shape as lesson 10's contract — three checks. SKELETON.
func runStoreContract(t *testing.T, s Store) {
	t.Helper()
	// TODO: implement the three contract checks (same as L10):
	//   1. Save then Load roundtrips.
	//   2. Save replaces (doesn't merge).
	// Note: we don't check "empty Load → empty + nil" here because
	// JSONStore now returns ErrNotFound for missing files. The
	// TestJSONStoreLoadReturnsErrNotFound test below covers that case.
	_ = s
	_ = expense.Expense{}
	_ = reflect.DeepEqual
}

// TestJSONStore is a SKELETON. Construct a JSONStore at a temp path,
// SAVE first (so the file exists), then run the shared contract.
func TestJSONStore(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "expenses.json")
	//   runStoreContract(t, NewJSONStore(path))
	_ = filepath.Join
	_ = t
}

// TestMemoryStore is a SKELETON. Same as TestJSONStore but no temp dir.
func TestMemoryStore(t *testing.T) {
	// TODO: runStoreContract(t, NewMemoryStore())
	_ = t
}

// TestJSONStoreLoadReturnsErrNotFound is a SKELETON. The lesson 11 upgrade:
// when the JSON file doesn't exist, JSONStore.Load returns a wrapped
// ErrNotFound. Verify errors.Is finds it.
//
// Cases to assert:
//   - err is non-nil
//   - errors.Is(err, ErrNotFound) is true
//   - err.Error() mentions the path (for human readability)
func TestJSONStoreLoadReturnsErrNotFound(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "does-not-exist.json")
	//   _, err := NewJSONStore(path).Load()
	//   assert err != nil; errors.Is(err, ErrNotFound); strings.Contains(err.Error(), path)
	_ = t
	_ = errors.Is
}

// TestJSONStoreLoadReturnsParseError is a SKELETON. Write a malformed
// JSON file to a temp path; JSONStore.Load should return a *ParseError
// extractable via errors.As, with Path == the temp path and Line > 0.
func TestJSONStoreLoadReturnsParseError(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "bad.json")
	//   os.WriteFile(path, []byte("not json"), 0o644)
	//   _, err := NewJSONStore(path).Load()
	//   var pe *ParseError
	//   if !errors.As(err, &pe) → t.Fatal("expected *ParseError")
	//   if pe.Path != path → t.Errorf("path: %q", pe.Path)
	//   if pe.Line == 0 → t.Errorf("expected non-zero Line")
	//   if pe.Cause == nil → t.Errorf("Cause is nil")
	_ = t
}

// TestMemoryStoreLoadNeverReturnsErrNotFound is a SKELETON. MemoryStore
// is not a file — empty IS valid. Verify Load on an empty MemoryStore
// returns no error.
func TestMemoryStoreLoadNeverReturnsErrNotFound(t *testing.T) {
	// TODO:
	//   _, err := NewMemoryStore().Load()
	//   if err != nil → t.Errorf("got %v, want nil", err)
	_ = t
}
