package store

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/11-errors/solutions/expense"
)

func runStoreContract(t *testing.T, s Store) {
	t.Helper()

	want := []expense.Expense{
		{Date: "2026-05-20", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-20", Amount: 12, Category: "lunch"},
		{Date: "2026-05-20", Amount: 75, Category: "rent"},
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load after Save: got %v, want %v", got, want)
	}

	replacement := []expense.Expense{
		{Date: "2026-05-20", Amount: 1.50, Category: "snack"},
	}
	if err := s.Save(replacement); err != nil {
		t.Fatalf("Save replacement: %v", err)
	}
	got, err = s.Load()
	if err != nil {
		t.Fatalf("Load after replacement: %v", err)
	}
	if !reflect.DeepEqual(got, replacement) {
		t.Errorf("Load after replacement: got %v, want %v", got, replacement)
	}
}

func TestJSONStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "expenses.json")
	runStoreContract(t, NewJSONStore(path))
}

func TestMemoryStore(t *testing.T) {
	runStoreContract(t, NewMemoryStore())
}

func TestJSONStoreLoadReturnsErrNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, err := NewJSONStore(path).Load()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false; err = %v", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error message %q doesn't mention path %q", err.Error(), path)
	}
}

func TestJSONStoreLoadReturnsParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	_, err := NewJSONStore(path).Load()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError; got %T: %v", err, err)
	}
	if pe.Path != path {
		t.Errorf("pe.Path = %q, want %q", pe.Path, path)
	}
	if pe.Line == 0 {
		t.Errorf("expected non-zero Line; got 0")
	}
	if pe.Cause == nil {
		t.Errorf("pe.Cause is nil")
	}
}

func TestMemoryStoreLoadNeverReturnsErrNotFound(t *testing.T) {
	_, err := NewMemoryStore().Load()
	if err != nil {
		t.Errorf("MemoryStore.Load returned error: %v", err)
	}
}
