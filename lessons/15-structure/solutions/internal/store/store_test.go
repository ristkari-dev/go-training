package store

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/15-structure/solutions/internal/expense"
)

var sample = []expense.Expense{
	{Date: "2026-05-21", Amount: 4.50, Category: "coffee"},
	{Date: "2026-05-21", Amount: 12.00, Category: "lunch"},
}

func runStoreContract(t *testing.T, name string, build func() Store) {
	t.Run(name+"/Save_then_Load_roundtrips", func(t *testing.T) {
		s := build()
		if err := s.Save(sample); err != nil {
			t.Fatalf("Save: %v", err)
		}
		got, err := s.Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if !reflect.DeepEqual(got, sample) {
			t.Errorf("roundtrip: got %+v, want %+v", got, sample)
		}
	})
	t.Run(name+"/Save_replaces_existing", func(t *testing.T) {
		s := build()
		if err := s.Save(sample); err != nil {
			t.Fatalf("first save: %v", err)
		}
		replacement := []expense.Expense{{Date: "2026-05-22", Amount: 1.00, Category: "snack"}}
		if err := s.Save(replacement); err != nil {
			t.Fatalf("second save: %v", err)
		}
		got, err := s.Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if !reflect.DeepEqual(got, replacement) {
			t.Errorf("replace: got %+v, want %+v", got, replacement)
		}
	})
}

func TestJSONStore(t *testing.T) {
	runStoreContract(t, "JSONStore", func() Store {
		return NewJSONStore(filepath.Join(t.TempDir(), "expenses.json"))
	})
}

func TestMemoryStore(t *testing.T) {
	runStoreContract(t, "MemoryStore", func() Store {
		return NewMemoryStore()
	})
}

func TestJSONStoreLoadReturnsErrNotFound(t *testing.T) {
	s := NewJSONStore(filepath.Join(t.TempDir(), "does-not-exist.json"))
	_, err := s.Load()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestJSONStoreLoadReturnsParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewJSONStore(path)
	_, err := s.Load()
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %v", err)
	}
	if pe.Path != path {
		t.Errorf("ParseError.Path = %q, want %q", pe.Path, path)
	}
	if pe.Line < 1 {
		t.Errorf("expected ParseError.Line >= 1, got %d", pe.Line)
	}
}

func TestMemoryStoreLoadNeverReturnsErrNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Load()
	if errors.Is(err, ErrNotFound) {
		t.Error("MemoryStore should never return ErrNotFound")
	}
}
