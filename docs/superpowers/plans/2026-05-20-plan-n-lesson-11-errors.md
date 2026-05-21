# Plan N — Lesson 11 (Errors) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 11 of Phase 2 — students learn the mature Go error-handling toolkit: sentinel errors (`var ErrXxx = errors.New(...)`), wrapping with `fmt.Errorf("...: %w", err)`, inspecting wrapped errors with `errors.Is`/`errors.As`, custom error types via struct + `Error() string` + `Unwrap()`, and the error-design philosophy ("does the caller need to discriminate? add a sentinel or type. Otherwise just wrap."). The main exercise upgrades L10's store package with `store.ErrNotFound` (replaces the silent empty-slice-on-missing-file behaviour) and `store.ParseError` (custom type wrapping JSON parse failures with path + line + cause). The cmd/expenses CLI is updated to use `errors.Is(err, store.ErrNotFound)` for friendly first-time-use and `errors.As` to extract `*ParseError` for a richer error message.

**Architecture:** Same per-lesson pattern as Plans D-M. Six tasks. Skeleton tests in warmup + main (Phase 2 default). Heavy-explanatory slide deck with **four** concept blocks (sentinel errors → wrapping with %w → errors.Is/As → custom error types).

**Tech Stack:** Go 1.23 stdlib only (`encoding/json`, `errors`, `fmt`, `io/fs`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `testing`). Reveal.js 5.1.0.

---

## Scope

After Plan N: lesson 11 is complete; `make test` green; cmd/expenses works end-to-end; both `errors.Is(err, store.ErrNotFound)` and `errors.As(err, &pe)` paths exercised.

### Design decisions (3 user-approved + 6 plan-recommended)

**User-approved via brainstorming:**

1. **JSONStore's "file not found" returns `ErrNotFound`** (no longer silent-empty). Caller uses `errors.Is(err, store.ErrNotFound)` to handle first-time use. Pedagogically tight — the sentinel + `errors.Is` lesson demonstrates itself on the exact case students already understand from L10.

2. **`ParseError` has `Cause error + Line int + Path string`.** Three fields — enough context for a useful error message ("parse expenses.json:5: invalid value at offset 42"), not so many that the type becomes a kitchen-sink.

3. **Full L10 cmd/expenses carry-forward** with upgraded error handling. Same 3 subcommands; same flag parsing; new error-handling helpers (`loadOrEmpty(s Store)` extracts the "missing-file is friendly" pattern into one place so cmdAdd/cmdList/cmdSummary all use it).

**Plan-recommended:**

4. **`ErrNotFound` is wrapped, not bare.** `JSONStore.Load` returns `fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)` — the wrapped error has the path for human-readable messages, but `errors.Is(err, ErrNotFound)` still works to detect the sentinel. Teaches "wrap to add context; the sentinel survives."

5. **`ParseError.Unwrap()` returns Cause.** Enables `errors.Is`/`errors.As` to chase through the wrap chain. Without `Unwrap`, `errors.As(err, &pe)` wouldn't find the underlying `*json.SyntaxError`.

6. **Line number computed from byte offset.** `json.SyntaxError` and `json.UnmarshalTypeError` both have `Offset int64`. A tiny `lineFromOffset(data, offset)` helper counts newlines. Demonstrates `errors.As` to extract typed stdlib errors.

7. **Warmup is `parseage`** (per Phase 2 design). `ParseAge(s string) (int, error)` wraps `strconv.Atoi`'s error with the offending input string. Tiny demo of `fmt.Errorf("...: %w", err)`. Tests assert both the message format AND that `errors.Is(err, strconv.ErrSyntax)` still finds the underlying sentinel.

8. **No new public APIs in `expense/`.** Expense package carries forward verbatim from L10. The error-handling lesson lives in store + cmd/expenses.

9. **`MemoryStore` doesn't return `ErrNotFound`.** Memory stores have no "file" — the empty state IS valid state. Only `JSONStore` returns ErrNotFound. The runStoreContract test demonstrates this divergence (calls `errors.Is(err, ErrNotFound)` on the first Load — JSONStore returns true, MemoryStore returns false, both are correct).

---

## Plans F-M lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages.
3. `→` arrow consistency.
4. Common-mistake content in README (4 per lesson, one per concept).
5. Slides + README written inline by controller.
6. `gofmt -w .` and `go vet ./...` mentions.
7. Skeleton tests pass vacuously.
8. Scaffold + restructure (delete 8 flat scaffolder files; create subpackage tree).

---

## File Structure

After Plan N (16 files total — same shape as L10):

```
lessons/11-errors/
├── README.md                                       (Task 5)
├── slides/
│   ├── index.html, slides.md, assets/.gitkeep      (Task 1 + Task 4)
├── exercises/
│   ├── warmup/parseage/
│   │   ├── parseage.go                             (Task 2)
│   │   └── parseage_test.go                        (Task 2 — SKELETON)
│   ├── expense/expense.go                          (Task 3 — verbatim from L10)
│   ├── store/
│   │   ├── store.go                                (Task 3 — adds ErrNotFound + ParseError + lineFromOffset)
│   │   └── store_test.go                           (Task 3 — SKELETON; tests both impls + error sentinels/types)
│   └── cmd/expenses/main.go                        (Task 3 — adds loadOrEmpty helper using errors.Is)
└── solutions/
    └── (mirrored structure)
```

---

## Conventions

- **Branch:** `feature/plan-n-lesson-11-errors`
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold + restructure

Same dance as Plans J/K/L/M.

- [ ] **Step 1-7:** `make new-lesson NAME=11-errors`; delete 8 flat files; verify 4-file tree (README + slides); commit `feat(lessons): scaffold lesson 11-errors with empty subpackage layout`.

---

## Task 2: Author the warm-up — `parseage` subpackage

`ParseAge(s string) (int, error)` wraps `strconv.Atoi`'s error with the offending input. Tiny but exercises three concepts at once: wrapping with `%w`, the still-inspectable underlying sentinel (`strconv.ErrSyntax`), and the `errors.Is` check.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/11-errors/{exercises,solutions}/warmup/parseage`

- [ ] **Step 2:** Create `lessons/11-errors/exercises/warmup/parseage/parseage.go`:

```go
// Package parseage is the lesson 11 warm-up: a tiny demo of error wrapping.
//
// ParseAge takes a string and returns an int. When the input isn't a valid
// integer, ParseAge returns an error that:
//   - Mentions the offending input in its message (for humans).
//   - Wraps the underlying strconv error (so errors.Is(err, strconv.ErrSyntax)
//     still works for callers that want to discriminate).
//
// Both pieces matter. A bare strconv.Atoi error tells you "syntax error" but
// not WHICH input was bad. A new errors.New("invalid age") loses the
// underlying sentinel that callers might check.
package parseage

import "strconv"

// ParseAge parses s as a non-negative integer age.
//
// On success, returns (age, nil).
//
// On failure, returns (0, wrapped error). The wrapped error mentions the
// offending input AND preserves the underlying strconv error via %w so
// errors.Is(err, strconv.ErrSyntax) returns true.
//
// Examples:
//
//	ParseAge("25")      → (25, nil)
//	ParseAge("")        → (0, error)  // errors.Is(err, strconv.ErrSyntax) == true
//	ParseAge("twenty")  → (0, error)  // same
//	ParseAge("-1")      → (-1, nil)   // we don't validate non-negative; the underlying Atoi succeeds
//
// Hint: strconv.Atoi(s); if err != nil return fmt.Errorf("parseage: invalid
// age %q: %w", s, err).
func ParseAge(s string) (int, error) {
	_ = strconv.Atoi
	panic("TODO: call strconv.Atoi; on error return fmt.Errorf with %q and %w wrapping")
}
```

- [ ] **Step 3:** Create `lessons/11-errors/exercises/warmup/parseage/parseage_test.go` (SKELETON):

```go
package parseage

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

// TestParseAgeValid is a SKELETON. Cover at least two cases where the
// input is a valid integer (positive, zero) and assert (age, nil err).
func TestParseAgeValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		// TODO: at least 3 cases. "25" → 25, "0" → 0, "-1" → -1, etc.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got, err := ParseAge(tc.in)
			//   if err != nil → t.Fatalf("unexpected error: %v", err)
			//   if got != tc.want → t.Errorf("...")
			_ = tc
		})
	}
}

// TestParseAgeInvalidWraps is a SKELETON. For invalid inputs, assert
// THREE things:
//   1. The error is non-nil.
//   2. The error MESSAGE contains the offending input (so humans can see
//      what was wrong). Use strings.Contains.
//   3. errors.Is(err, strconv.ErrSyntax) is true — the underlying
//      strconv sentinel survives the wrap.
//
// Cases: empty string, non-numeric "twenty", spaces "  25  ".
func TestParseAgeInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		// TODO: at least 3 cases.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   _, err := ParseAge(tc.in)
			//   if err == nil → t.Fatalf("expected error for %q", tc.in)
			//   if !strings.Contains(err.Error(), tc.in) → t.Errorf("message missing input")
			//   if !errors.Is(err, strconv.ErrSyntax) → t.Errorf("wrapped sentinel not findable")
			_ = tc
			_ = errors.Is
			_ = strconv.ErrSyntax
			_ = strings.Contains
		})
	}
}
```

- [ ] **Step 4:** Create `lessons/11-errors/solutions/warmup/parseage/parseage.go`:

```go
// Package parseage is the lesson 11 warm-up reference implementation.
package parseage

import (
	"fmt"
	"strconv"
)

// ParseAge parses s as an integer age. Wraps strconv.Atoi errors with the
// offending input string.
func ParseAge(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parseage: invalid age %q: %w", s, err)
	}
	return n, nil
}
```

- [ ] **Step 5:** Create `lessons/11-errors/solutions/warmup/parseage/parseage_test.go`:

```go
package parseage

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestParseAgeValid(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"positive", "25", 25},
		{"zero", "0", 0},
		{"negative", "-1", -1},
		{"large", "1000000", 1000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseAge(tc.in)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseAge(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseAgeInvalidWraps(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"non-numeric", "twenty"},
		{"trailing-space", "25 "},
		{"with-letters", "25a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseAge(tc.in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
			if !strings.Contains(err.Error(), tc.in) {
				t.Errorf("error message %q doesn't mention input %q", err.Error(), tc.in)
			}
			if !errors.Is(err, strconv.ErrSyntax) {
				t.Errorf("err = %v; expected errors.Is(err, strconv.ErrSyntax) to be true", err)
			}
		})
	}
}
```

> Note on `"-1"`: a negative "age" doesn't make real-world sense, but `strconv.Atoi("-1")` returns `-1, nil` (it's a valid integer). ParseAge doesn't validate semantically — only syntactically. Lesson 11 isn't about validation; it's about error wrapping.

> Note on the empty case: `strconv.Atoi("")` returns `0, *strconv.NumError{Func: "Atoi", Num: "", Err: strconv.ErrSyntax}`. The `Err` field is `ErrSyntax`. `errors.Is(numErr, strconv.ErrSyntax)` works because `*strconv.NumError` has an `Unwrap()` method returning its `Err`. Then our wrap with `%w` makes `errors.Is(ourErr, strconv.ErrSyntax)` work through both layers.

- [ ] **Step 6:** Verify (gofmt + tests + lint + vet); commit:

```bash
gofmt -l lessons/11-errors/
go test ./lessons/11-errors/exercises/warmup/parseage/... -v 2>&1 | tail -10
go test ./lessons/11-errors/solutions/warmup/parseage/... -v 2>&1 | tail -15
make test
golangci-lint run ./...
go vet ./...

git add lessons/11-errors/exercises/warmup/ lessons/11-errors/solutions/warmup/
git commit -m "feat(lesson-11): warmup — parseage (wrap strconv.Atoi with %w)"
```

Expected: gofmt empty; exercises pass vacuously; solutions TestParseAgeValid (4 sub-tests) + TestParseAgeInvalidWraps (4 sub-tests) pass; make test green; lint 0 issues; vet clean.

---

## Task 3: Main — `store` (with sentinels + custom error type) + carry-forward `expense` + updated `cmd/expenses`

The biggest task — three subpackages, the cmd binary, the new error machinery in store.

**Files (12 total — 6 per side):**

- [ ] **Step 1:** Create directories.

```bash
mkdir -p lessons/11-errors/{exercises,solutions}/{expense,store,cmd/expenses}
```

- [ ] **Step 2:** Create `lessons/11-errors/exercises/expense/expense.go` (verbatim L10 forward-port — same as Plan M Step 2). FULLY IMPLEMENTED. Same content:

```go
// Package expense holds the Expense type for the lesson 11 capstone.
//
// Carried forward verbatim from lesson 10 (4 methods: Format, IsHigh,
// ApplyDiscount, Bump; JSON tags intact). No test in this lesson —
// already tested in L09. Lesson 11's store and cmd/expenses packages
// import this.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
```

- [ ] **Step 3:** Same content in `lessons/11-errors/solutions/expense/expense.go`.

- [ ] **Step 4:** Create `lessons/11-errors/exercises/store/store.go`:

```go
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
//   1. data, err := os.ReadFile(j.Path)
//   2. if errors.Is(err, fs.ErrNotExist) → return nil, fmt.Errorf("loading %s: %w", j.Path, ErrNotFound)
//   3. if err != nil → return nil, fmt.Errorf("loading %s: %w", j.Path, err)
//   4. var es []expense.Expense
//   5. if err := json.Unmarshal(data, &es); err != nil → return nil,
//        &ParseError{Path: j.Path, Line: lineFromJSONErr(data, err), Cause: err}
//   6. if es == nil → es = []expense.Expense{}
//   7. return es, nil
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
//
// Hint: errors.As to extract *json.SyntaxError or *json.UnmarshalTypeError
// (both have an Offset int64 field); pass the offset to lineFromOffset.
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
var _ = filepath.Join
var _ = fs.ErrNotExist
var _ = json.Unmarshal
var _ = os.ReadFile
var _ = errors.Is
var _ = fmt.Errorf
```

> Note: `MemoryStore` ships fully implemented (carried from L10) because the lesson's focus is the error-handling upgrade in `JSONStore` + `ParseError`/`ErrNotFound`. Students implement the error-related bits; the existing logic doesn't need re-implementing.

- [ ] **Step 5:** Create `lessons/11-errors/exercises/store/store_test.go` (SKELETON):

```go
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
	//   1. Load on empty store returns non-nil empty slice + nil error.
	//   2. Save then Load roundtrips.
	//   3. Save replaces (doesn't merge).
	_ = s
	_ = expense.Expense{}
	_ = reflect.DeepEqual
}

// TestJSONStore is a SKELETON. Construct a JSONStore at a temp path,
// run the shared contract.
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
```

- [ ] **Step 6:** Create `lessons/11-errors/solutions/store/store.go`:

```go
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
```

- [ ] **Step 7:** Create `lessons/11-errors/solutions/store/store_test.go`:

```go
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

	// On a fresh store, Load may return either ([], nil) (MemoryStore)
	// or ErrNotFound (JSONStore at a non-existent path). For the contract
	// we only check that subsequent Save→Load roundtrips work.

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
```

> Note: the lesson 11 `runStoreContract` is slightly different from L10's — the "empty Load returns non-nil empty slice + nil error" check is gone because JSONStore.Load now returns ErrNotFound for empty/missing. The TestJSONStoreLoadReturnsErrNotFound test exercises the new behaviour explicitly.

- [ ] **Step 8:** Create `lessons/11-errors/exercises/cmd/expenses/main.go` (FULL WORKING CODE — students study; same shape as L10 with new `loadOrEmpty` helper):

```go
// Package main is the lesson 11 expense tracker CLI.
//
// Same as lesson 10 + upgraded error handling:
//   - Uses errors.Is(err, store.ErrNotFound) to make first-time use friendly.
//   - Uses errors.As to extract *store.ParseError for richer error messages
//     when the JSON file is malformed.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/11-errors/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/11-errors/exercises/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		var pe *store.ParseError
		if errors.As(err, &pe) {
			fmt.Fprintf(os.Stderr, "error: could not parse %s at line %d: %v\n",
				pe.Path, pe.Line, pe.Cause)
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	storeKind, args, err := parseStoreOption(args)
	if err != nil {
		return err
	}
	path, args, err := parseFileOption(args)
	if err != nil {
		return err
	}

	s, err := buildStore(storeKind, path)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing subcommand (try: add, list, summary)")
	}
	switch args[0] {
	case "add":
		return cmdAdd(s, args[1:])
	case "list":
		return cmdList(s)
	case "summary":
		return cmdSummary(s)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseStoreOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-store=") {
		v := args[0][len("-store="):]
		if v != "mem" && v != "json" {
			return "", nil, fmt.Errorf("invalid -store value %q (must be mem or json)", v)
		}
		return v, args[1:], nil
	}
	return "json", args, nil
}

func parseFileOption(args []string) (string, []string, error) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-file=") {
		return args[0][len("-file="):], args[1:], nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".expenses.json"), args, nil
}

func buildStore(kind, path string) (store.Store, error) {
	switch kind {
	case "mem":
		return store.NewMemoryStore(), nil
	case "json":
		return store.NewJSONStore(path), nil
	default:
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

// loadOrEmpty calls s.Load and treats ErrNotFound as "empty + nil" so
// callers don't need to repeat the errors.Is check in every subcommand.
// Other errors pass through.
func loadOrEmpty(s store.Store) ([]expense.Expense, error) {
	es, err := s.Load()
	if errors.Is(err, store.ErrNotFound) {
		return []expense.Expense{}, nil
	}
	return es, err
}

func cmdAdd(s store.Store, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	e := expense.Expense{Date: args[0], Amount: amount, Category: args[2]}
	es = append(es, e)
	if err := s.Save(es); err != nil {
		return err
	}
	fmt.Println("added:", e.Format())
	return nil
}

func cmdList(s store.Store) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(s store.Store) error {
	es, err := loadOrEmpty(s)
	if err != nil {
		return err
	}

	var total float64
	for _, e := range es {
		total += e.Amount
	}
	fmt.Printf("%d expenses, total €%.2f\n", len(es), total)
	if len(es) == 0 {
		return nil
	}

	fmt.Println("by category:")
	totals := map[string]float64{}
	for _, e := range es {
		totals[e.Category] += e.Amount
	}

	keys := make([]string, 0, len(totals))
	for k := range totals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-10s €%-7.2f\n", k, totals[k])
	}
	return nil
}
```

- [ ] **Step 9:** Same content in `lessons/11-errors/solutions/cmd/expenses/main.go` but importing from `solutions/expense` and `solutions/store`.

- [ ] **Step 10:** Verify (gofmt + tests + lint + vet + smoke tests):

```bash
gofmt -l lessons/11-errors/
go test ./lessons/11-errors/exercises/... -v 2>&1 | tail -30
go test ./lessons/11-errors/solutions/... -v 2>&1 | tail -40
make test
golangci-lint run ./...
go vet ./...

# Smoke test the solutions binary
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP
echo "--- first-time use (file missing — ErrNotFound path) ---"
go run ./lessons/11-errors/solutions/cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
echo "--- normal list ---"
go run ./lessons/11-errors/solutions/cmd/expenses -store=json -file=$TMP list

echo "--- malformed JSON (ParseError path) ---"
echo "not json" > $TMP
go run ./lessons/11-errors/solutions/cmd/expenses -store=json -file=$TMP list 2>&1 || echo "(exited 1 as expected)"

rm $TMP
```

Expected:
- Tests pass (exercises vacuous, solutions all green)
- Lint 0; vet clean
- First-time use prints `added: ...` (no error message, despite the file not existing — loadOrEmpty handled it)
- List prints the one expense
- Malformed JSON prints `error: could not parse /tmp/expenses-XXXXXX.json at line 1: invalid character 'o' in literal null (expecting 'u')` (or similar) and exits 1

- [ ] **Step 11:** Commit:

```bash
git add lessons/11-errors/exercises/expense/ lessons/11-errors/exercises/store/ lessons/11-errors/exercises/cmd/ \
        lessons/11-errors/solutions/expense/ lessons/11-errors/solutions/store/ lessons/11-errors/solutions/cmd/
git commit -m "feat(lesson-11): main — ErrNotFound sentinel + ParseError custom type + loadOrEmpty helper"
```

---

## Task 4: Slide deck (controller, inline)

4 concepts: sentinel errors → wrapping with `%w` → `errors.Is`/`errors.As` → custom error types. Same structural pattern (title slide, "What we'll cover", 4× concept with Motivation/Basics/Worked example/Common mistake/Recap, Practice, What we learned, Up next).

Commit: `feat(lesson-11): slides — Errors (4 concepts)`.

---

## Task 5: README (controller, inline)

Mirrors slide concepts; Common-mistake examples per concept; "What's different" note about the L10 → L11 behaviour change in JSONStore.

Commit: `docs(lesson-11): README — Errors self-study`.

---

## Task 6: End-to-end verification

`make test`; `make test-lesson LESSON=11-errors`; lint + vet; slides-build; binary smoke tests (both `-store=json` first-time + listing, and `-store=mem`). No commit.

---

## Done definition

- 16 files in lesson 11.
- `make test`, `golangci-lint`, `go vet` all green.
- Both `-store=json` (first-time + normal + malformed file) and `-store=mem` smoke tests work end-to-end.
- 6 commits on the branch.

## What ships next

**Plan O — Lesson 12 (Generics, standalone).** Type parameters, constraints (`cmp.Ordered`), when NOT to use generics. Standalone example: a small `slices`-style helper library (Filter/Map/Max). Tracker doesn't evolve this lesson — forcing generics into the tracker would be the premature-abstraction anti-pattern Go warns against. New import: `cmp`.
