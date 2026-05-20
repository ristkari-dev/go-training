# Plan M — Lesson 10 (Interfaces) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 10 of Phase 2 — students learn implicit interface satisfaction, small interfaces (`io.Reader`/`io.Writer`/`error`), `any`, type assertions, and the "accept interfaces, return structs" rule. The main exercise introduces a `Store` interface to the expense tracker — abstracts persistence — with two implementations (`JSONStore` carrying forward lesson 08's storage logic, plus a new `MemoryStore`). The CLI gets a `-store=mem|json` flag so the runtime can pick.

**Architecture:** Same per-lesson pattern as Plans D-L. Six tasks: scaffold + restructure, warmup, main (expense + store + cmd/expenses together), slides, README, end-to-end verify. Subpackages + `cmd/` (lesson 10 has a binary — the refactored CLI). Skeleton tests in both warmup and main (Phase 2 default). Heavy-explanatory slide deck with **four** concept blocks (implicit satisfaction → small interfaces → `any` + type assertions → interface design).

**Tech Stack:** Go 1.23 stdlib only (`encoding/json`, `errors`, `fmt`, `io/fs`, `os`, `path/filepath`, `sort`, `strconv`, `strings`, `testing`). Reveal.js 5.1.0 for the deck.

---

## Scope

Plan M produces lesson 10 only. After Plan M lands:

- `lessons/10-interfaces/` is a complete teachable lesson with all four parts populated.
- `make test` passes; `make test-exercises` shows lesson 10's tests passing vacuously.
- `make slides-dev LESSON=10-interfaces` serves the deck.
- The lesson is included in `dist/index.html`.
- `go run ./lessons/10-interfaces/solutions/cmd/expenses -store=json -file=$TMP add ...` works end-to-end; `-store=mem` works in-process.

**Out of scope (handled by future plans):** Lessons 11-15.

### Design decisions made during planning

Three calls were locked in via brainstorming with the user; six more are this plan's recommendations.

1. **`Store` interface is whole-blob** (per user): `Load() ([]expense.Expense, error)` + `Save([]expense.Expense) error`. Mirrors lesson 08's `storage.LoadExpenses`/`SaveExpenses`. Keeps focus on interface mechanics, not expense-domain redesign.

2. **Plain `store/` subpackage** (per user, not `internal/store/`). `internal/` formally arrives in L15.

3. **Full L08 CLI shape preserved** (per user): add/list/summary subcommands stay; the binary is refactored to use `store.Store` interface. Cmd ships as working starter code; the exercise is implementing `Store` + `JSONStore` + `MemoryStore`.

4. **Two warmup implementations: `English` and `Finnish`.** Smallest pedagogically clean way to demonstrate "two unrelated types satisfy the same interface." Greeting in two languages is culturally neutral and obvious.

5. **`NewJSONStore` and `NewMemoryStore` constructors.** Returning `*Store` would be wrong (anti-pattern: pointers to interfaces are almost never what you want). Returning the concrete types directly (`*JSONStore`, `*MemoryStore`) is the canonical Go idiom — callers assign to a `store.Store` interface variable when they want polymorphism.

6. **Pointer receivers on JSONStore + MemoryStore.** `JSONStore` doesn't mutate state (the file is the state), but consistency wins — both stores use pointer receivers across all methods. Matches the lesson 09 "be consistent across a type's methods" guidance.

7. **`MemoryStore` is concurrency-naive.** No mutex; the lesson is about interfaces, not concurrency. Lesson 18 (sync) returns to MemoryStore and adds the mutex.

8. **`-store=mem` won't persist across invocations.** That's by design — `MemoryStore` is for testing. The slide deck and README explicitly flag this so students aren't surprised when `mem add ...` followed by `mem list` returns empty.

9. **Forward-port `expense/expense.go` verbatim from L09.** Same struct, same four methods, same JSON tags. No test file in L10 (already tested in L09). The `store` subpackage's tests use `expense.Expense` extensively; that exercises the type indirectly.

---

## Plans F-L lessons-learned applied here

1. **Subpackages have their own namespace** — no `Warmup*` prefix needed (continued from L09).

2. **Lint exclusion covers nested subpackages** — `.golangci.yml`'s `lessons/.*/exercises/` regex matches `lessons/10-interfaces/exercises/store/store.go`.

3. **`→` arrow consistency** — Unicode arrows in doc-comment examples; gofmt may reformat blocks; arrows survive.

4. **Common-mistake content in README** — every slide concept's common-mistake example is mirrored in the README. 4 total.

5. **Slides + README written inline by controller** — Plans G-L established this; Plan M continues. Tasks 1-3 + 6 can be subagent-driven; Tasks 4-5 the controller writes directly.

6. **`gofmt -w .` and `go vet ./...` mentions** — README continues.

7. **Format-output verification** — Lesson 10 doesn't re-test `Expense.Format()` (forward-port is byte-identical with L09).

8. **Empty case → vacuous pass.** Skeleton tests pass vacuously until students add cases.

9. **Scaffold + restructure** — same dance as Plans J/K/L (delete 8 flat scaffolder files, create subpackage tree).

---

## File Structure

After Plan M (16 files total):

```
lessons/10-interfaces/
├── README.md                                       (Task 5)
├── slides/
│   ├── index.html                                  (Task 1; unchanged from scaffold)
│   ├── slides.md                                   (Task 4 — controller writes inline)
│   └── assets/.gitkeep                             (Task 1)
├── exercises/
│   ├── warmup/
│   │   └── greet/
│   │       ├── greet.go                            (Task 2 — Greeter + English + Finnish stubs)
│   │       └── greet_test.go                       (Task 2 — SKELETON)
│   ├── expense/
│   │   └── expense.go                              (Task 3 — verbatim forward-port from L09)
│   ├── store/
│   │   ├── store.go                                (Task 3 — Store interface + JSONStore + MemoryStore stubs)
│   │   └── store_test.go                           (Task 3 — SKELETON for both impls)
│   └── cmd/
│       └── expenses/
│           └── main.go                             (Task 3 — refactored L08 CLI; working code)
└── solutions/
    ├── warmup/
    │   └── greet/
    │       ├── greet.go                            (Task 2 — implementation)
    │       └── greet_test.go                       (Task 2 — full reference)
    ├── expense/
    │   └── expense.go                              (Task 3 — same as L09 solutions/expense)
    ├── store/
    │   ├── store.go                                (Task 3 — implementation)
    │   └── store_test.go                           (Task 3 — full reference)
    └── cmd/
        └── expenses/
            └── main.go                             (Task 3 — implementation)
```

Files deleted from the scaffolder output (8): same as Plans J/K/L (the 4 flat `.go` files per side).

### Decomposition rationale

Task 3 combines main exercise (expense + store + cmd) into one task because the three are tightly coupled (cmd imports store + expense; store imports expense). Single subagent dispatch keeps the context coherent. Plan K (L08 capstone) split this into 4-5 tasks because of size; L10 is smaller (no integration tests; no summary subpackage; no provided storage package).

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/plan-m-lesson-10-interfaces`
- **Commit messages:** Conventional Commits

---

## Task 1: Scaffold + restructure

**Files:**
- Create via scaffolder: 12 flat files
- Delete: 8 `.go` files at top level

- [ ] **Step 1:** Confirm branch is `feature/plan-m-lesson-10-interfaces`, one commit ahead of `main`.

- [ ] **Step 2:** `make new-lesson NAME=10-interfaces`

- [ ] **Step 3:** Delete the 8 flat `.go` files:

```bash
rm lessons/10-interfaces/exercises/main.go \
   lessons/10-interfaces/exercises/main_test.go \
   lessons/10-interfaces/exercises/warmup.go \
   lessons/10-interfaces/exercises/warmup_test.go \
   lessons/10-interfaces/solutions/main.go \
   lessons/10-interfaces/solutions/main_test.go \
   lessons/10-interfaces/solutions/warmup.go \
   lessons/10-interfaces/solutions/warmup_test.go
```

- [ ] **Step 4:** Verify tree (4 files: README + slides/{index.html, slides.md, assets/.gitkeep}).

- [ ] **Step 5:** `make test` passes.

- [ ] **Step 6:** Commit:

```bash
git add -A lessons/10-interfaces/
git commit -m "feat(lessons): scaffold lesson 10-interfaces with empty subpackage layout"
```

---

## Task 2: Author the warm-up — `greet` subpackage

A `Greeter` interface with one method, and two unrelated struct types that both satisfy it. Smallest possible "implicit satisfaction" demo.

**Files:**
- Create: `lessons/10-interfaces/exercises/warmup/greet/greet.go`
- Create: `lessons/10-interfaces/exercises/warmup/greet/greet_test.go`
- Create: `lessons/10-interfaces/solutions/warmup/greet/greet.go`
- Create: `lessons/10-interfaces/solutions/warmup/greet/greet_test.go`

- [ ] **Step 1:** Create directories.

```bash
mkdir -p lessons/10-interfaces/exercises/warmup/greet \
         lessons/10-interfaces/solutions/warmup/greet
```

- [ ] **Step 2:** Create `lessons/10-interfaces/exercises/warmup/greet/greet.go`:

```go
// Package greet is the lesson 10 warm-up: a tiny demo of interface
// satisfaction. The Greeter interface has one method; two struct types
// (English, Finnish) each implement it — and that's all it takes for them
// to be "Greeters." No `implements` keyword. No declared relationship.
package greet

// Greeter is a one-method interface: anything with `Greet(name string) string`
// is a Greeter.
type Greeter interface {
	Greet(name string) string
}

// English greets in English.
type English struct{}

// Greet returns "Hello, <name>!".
//
// Examples:
//
//	English{}.Greet("World")  → "Hello, World!"
//	English{}.Greet("Aki")    → "Hello, Aki!"
//
// Hint: fmt.Sprintf("Hello, %s!", name).
func (English) Greet(name string) string {
	panic("TODO: return \"Hello, <name>!\"")
}

// Finnish greets in Finnish.
type Finnish struct{}

// Greet returns "Hei, <name>!".
//
// Examples:
//
//	Finnish{}.Greet("World")  → "Hei, World!"
//	Finnish{}.Greet("Aki")    → "Hei, Aki!"
//
// Hint: fmt.Sprintf("Hei, %s!", name).
func (Finnish) Greet(name string) string {
	panic("TODO: return \"Hei, <name>!\"")
}
```

> Note on `func (English) Greet(...)`: the receiver is unnamed because the method body doesn't use it (it's a "marker" method on a stateless struct). Go allows this. Same shape for `Finnish`. The doc comments on the methods substitute for the receiver names.

- [ ] **Step 3:** Create `lessons/10-interfaces/exercises/warmup/greet/greet_test.go` (SKELETON):

```go
package greet

import (
	"strings"
	"testing"
)

// TestGreeters is a SKELETON. The point of the test is to demonstrate
// that English and Finnish BOTH satisfy the Greeter interface — without
// either type declaring that it does. We loop over `[]Greeter{...}` and
// call Greet on each.
//
// Cases to add: at minimum, English with "World" → "Hello, World!", and
// Finnish with "World" → "Hei, World!". Use strings.HasPrefix for a
// language-agnostic shape check, OR direct equality for exact match —
// your choice.
func TestGreeters(t *testing.T) {
	cases := []struct {
		name string
		g    Greeter
		in   string
		want string
	}{
		// TODO: at least 4 cases. English with "World", English with
		// "Aki", Finnish with "World", Finnish with empty string.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: got := tc.g.Greet(tc.in)
			//   if got != tc.want { t.Errorf(...) }
			_ = tc
			_ = strings.HasPrefix
		})
	}
}

// TestGreeterSlice is a SKELETON. Verify that a `[]Greeter` slice
// containing both an English and a Finnish value works as one
// homogeneous collection — that's the payoff of implicit satisfaction.
func TestGreeterSlice(t *testing.T) {
	// TODO:
	//   greeters := []Greeter{English{}, Finnish{}}
	//   for _, g := range greeters {
	//       got := g.Greet("World")
	//       if got == "" { t.Error("empty greeting") }
	//   }
	_ = t
}
```

- [ ] **Step 4:** Create `lessons/10-interfaces/solutions/warmup/greet/greet.go`:

```go
// Package greet is the lesson 10 warm-up reference implementation.
package greet

import "fmt"

// Greeter is a one-method interface.
type Greeter interface {
	Greet(name string) string
}

// English greets in English.
type English struct{}

// Greet returns "Hello, <name>!".
func (English) Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// Finnish greets in Finnish.
type Finnish struct{}

// Greet returns "Hei, <name>!".
func (Finnish) Greet(name string) string {
	return fmt.Sprintf("Hei, %s!", name)
}
```

- [ ] **Step 5:** Create `lessons/10-interfaces/solutions/warmup/greet/greet_test.go`:

```go
package greet

import "testing"

func TestGreeters(t *testing.T) {
	cases := []struct {
		name string
		g    Greeter
		in   string
		want string
	}{
		{"english-world", English{}, "World", "Hello, World!"},
		{"english-name", English{}, "Aki", "Hello, Aki!"},
		{"finnish-world", Finnish{}, "World", "Hei, World!"},
		{"finnish-name", Finnish{}, "Aki", "Hei, Aki!"},
		{"english-empty", English{}, "", "Hello, !"},
		{"finnish-empty", Finnish{}, "", "Hei, !"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.g.Greet(tc.in); got != tc.want {
				t.Errorf("%T.Greet(%q) = %q, want %q", tc.g, tc.in, got, tc.want)
			}
		})
	}
}

func TestGreeterSlice(t *testing.T) {
	greeters := []Greeter{English{}, Finnish{}}
	for _, g := range greeters {
		got := g.Greet("World")
		if got == "" {
			t.Errorf("%T.Greet returned empty string", g)
		}
	}
}
```

- [ ] **Step 6:** gofmt, run tests, lint, vet:

```bash
gofmt -l lessons/10-interfaces/
go test ./lessons/10-interfaces/exercises/warmup/greet/... -v 2>&1 | tail -10
go test ./lessons/10-interfaces/solutions/warmup/greet/... -v 2>&1 | tail -15
make test
golangci-lint run ./...
go vet ./...
```

Expected: gofmt empty; exercises pass vacuously; solutions TestGreeters (6 sub-tests) + TestGreeterSlice pass; make test green; lint 0 issues; vet clean.

- [ ] **Step 7:** Commit:

```bash
git add lessons/10-interfaces/exercises/warmup/ lessons/10-interfaces/solutions/warmup/
git commit -m "feat(lesson-10): warmup — greet (Greeter interface + English + Finnish impls)"
```

---

## Task 3: Author the main exercise — `expense` + `store` + `cmd/expenses`

Three subpackages in one task because they're tightly coupled:
- `expense/expense.go` — forward-port from L09 verbatim
- `store/store.go` — Store interface + JSONStore (file-backed) + MemoryStore (in-process)
- `cmd/expenses/main.go` — refactored L08 CLI with `-store=mem|json` flag

**Files:**
- Create: 6 files per side = 12 files

- [ ] **Step 1:** Create directories.

```bash
mkdir -p lessons/10-interfaces/exercises/{expense,store,cmd/expenses} \
         lessons/10-interfaces/solutions/{expense,store,cmd/expenses}
```

- [ ] **Step 2:** Create `lessons/10-interfaces/exercises/expense/expense.go` (verbatim forward-port from L09):

```go
// Package expense holds the Expense type for the lesson 10 capstone.
//
// Carried forward verbatim from lesson 09 (4 methods: Format, IsHigh,
// ApplyDiscount, Bump; JSON tags intact). No test in this lesson —
// already tested in L09. Lesson 10's store and cmd/expenses packages
// import this.
package expense

import "fmt"

// Expense is one row of the expense tracker.
type Expense struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

// Format returns "YYYY-MM-DD  €AMOUNT  category" with "%s  €%-7.2f %s".
func (e Expense) Format() string {
	return fmt.Sprintf("%s  €%-7.2f %s", e.Date, e.Amount, e.Category)
}

// IsHigh reports whether the expense is over €50.
func (e Expense) IsHigh() bool {
	return e.Amount > 50
}

// ApplyDiscount reduces e.Amount by the given rate. Pointer receiver.
func (e *Expense) ApplyDiscount(rate float64) {
	e.Amount *= (1 - rate)
}

// Bump adds amount to e.Amount. Pointer receiver.
func (e *Expense) Bump(amount float64) {
	e.Amount += amount
}
```

> Note: this file ships **fully implemented** in both exercises/ and solutions/. The lesson 10 exercises don't ask students to re-implement Expense — they need the type to use Store. This is one of the few cases where exercises/ doesn't have a panic-stub.

- [ ] **Step 3:** Create the same file for solutions/ (identical to step 2).

- [ ] **Step 4:** Create `lessons/10-interfaces/exercises/store/store.go`:

```go
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
	panic("TODO: read j.Path; on fs.ErrNotExist return ([], nil); else json.Unmarshal")
}

// Save writes es to j.Path as pretty-printed JSON.
//
// Hint: json.MarshalIndent(es, "", "  "); os.MkdirAll for the parent
// directory if it doesn't exist; os.WriteFile to write the bytes.
func (j *JSONStore) Save(es []expense.Expense) error {
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
```

- [ ] **Step 5:** Create `lessons/10-interfaces/exercises/store/store_test.go` (SKELETON):

```go
package store

import (
	"path/filepath"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/expense"
)

// runStoreContract is a helper that verifies the Store CONTRACT — i.e.,
// behaviours every implementation must support. Both TestJSONStore and
// TestMemoryStore call it.
//
// SKELETON. Fill in the body to:
//   1. Load() on an empty store returns ([]expense.Expense{}, nil) — non-nil
//      empty slice, nil error.
//   2. After Save([e1, e2, e3]), Load() returns those three in order.
//   3. After a second Save([e4]), Load() returns just [e4] (Save REPLACES).
//
// Hint: define a fixture once at the top of the function:
//
//	e1 := expense.Expense{Date: "2026-05-20", Amount: 4.50, Category: "coffee"}
//	...
func runStoreContract(t *testing.T, s Store) {
	t.Helper()
	// TODO: implement the three contract checks described above.
	_ = s
}

// TestJSONStore is a SKELETON. Construct a JSONStore at a temp path, then
// hand it to runStoreContract. The temp path uses t.TempDir() so the test
// is hermetic and self-cleaning.
func TestJSONStore(t *testing.T) {
	// TODO:
	//   path := filepath.Join(t.TempDir(), "expenses.json")
	//   runStoreContract(t, NewJSONStore(path))
	_ = filepath.Join
	_ = t
}

// TestMemoryStore is a SKELETON. Construct a fresh MemoryStore and hand
// it to runStoreContract. No temp dir needed — memory is process-local.
func TestMemoryStore(t *testing.T) {
	// TODO: runStoreContract(t, NewMemoryStore())
	_ = t
}
```

- [ ] **Step 6:** Create `lessons/10-interfaces/solutions/store/store.go`:

```go
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
```

- [ ] **Step 7:** Create `lessons/10-interfaces/solutions/store/store_test.go`:

```go
package store

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/solutions/expense"
)

// runStoreContract verifies behaviours every Store implementation must support.
func runStoreContract(t *testing.T, s Store) {
	t.Helper()

	// 1. Load on an empty store returns a non-nil empty slice and nil error.
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if got == nil {
		t.Errorf("Load empty: got nil slice, want non-nil empty")
	}
	if len(got) != 0 {
		t.Errorf("Load empty: got %d items, want 0", len(got))
	}

	// 2. Save then Load returns the same expenses in order.
	want := []expense.Expense{
		{Date: "2026-05-20", Amount: 4.50, Category: "coffee"},
		{Date: "2026-05-20", Amount: 12, Category: "lunch"},
		{Date: "2026-05-20", Amount: 75, Category: "rent"},
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err = s.Load()
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load after Save: got %v, want %v", got, want)
	}

	// 3. A second Save REPLACES (doesn't merge).
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
```

> Note on `runStoreContract`: same test logic verifies both implementations satisfy the Store contract. This is *the* idiomatic Go pattern for testing interface implementations — write the contract once, hand each impl to it. Lesson 15's `testing patterns` section returns to this pattern formally.

- [ ] **Step 8:** Create `lessons/10-interfaces/exercises/cmd/expenses/main.go` (FULL WORKING CODE — students study this, don't reimplement):

```go
// Package main is the lesson 10 expense tracker CLI — same shape as L08's
// CLI but refactored to use the store.Store interface, with a new
// -store=mem|json flag to pick the implementation at runtime.
//
// Usage:
//
//	expenses [-store=mem|json] [-file=path] <subcommand> [args...]
//
// Subcommands:
//
//	add DATE AMOUNT CATEGORY    Append an expense.
//	list                         Print all expenses.
//	summary                      Print per-category totals.
//
// The -store flag selects the backing store: "json" (default) uses
// JSONStore at the -file path; "mem" uses MemoryStore (data is lost
// when the process exits — useful for tests).
//
// The -file flag (only meaningful for -store=json) defaults to
// $HOME/.expenses.json. Both flags must appear BEFORE the subcommand.
//
// IMPORTANT: This file ships fully working in lesson 10 — your job is
// to implement store.Store + JSONStore + MemoryStore in store/store.go.
// Once those are done, this binary works.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/expense"
	"github.com/ristkari-dev/go-training/lessons/10-interfaces/exercises/store"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
		// parseStoreOption guards this; defensive only.
		return nil, fmt.Errorf("unsupported store kind %q", kind)
	}
}

func cmdAdd(s store.Store, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: add DATE AMOUNT CATEGORY")
	}
	amount, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", args[1], err)
	}
	es, err := s.Load()
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
	es, err := s.Load()
	if err != nil {
		return err
	}
	for _, e := range es {
		fmt.Println(e.Format())
	}
	return nil
}

func cmdSummary(s store.Store) error {
	es, err := s.Load()
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

> Note: lesson 10 drops the bar-chart from cmdSummary (that lives in L08's `summary.Bar`; bringing the whole summary package over for the visual would distract from the interface lesson). The output is text-only per-category totals. Students see the bar chart again in L15's tracker reorg.

- [ ] **Step 9:** Create `lessons/10-interfaces/solutions/cmd/expenses/main.go` (same as Step 8 but imports from `.../solutions/expense` and `.../solutions/store`).

- [ ] **Step 10:** Format-output smoke test (carries from L09):

```bash
cat > /tmp/format-check.go <<'EOF'
package main
import "fmt"
func main() {
    fmt.Printf("%q\n", fmt.Sprintf("%s  €%-7.2f %s", "2026-05-20", 4.50, "coffee"))
}
EOF
go run /tmp/format-check.go
rm /tmp/format-check.go
```

Expected: `"2026-05-20  €4.50    coffee"` (byte-match for L09/L10's Format()).

- [ ] **Step 11:** gofmt, tests, lint, vet:

```bash
gofmt -l lessons/10-interfaces/
go test ./lessons/10-interfaces/exercises/... -v 2>&1 | tail -20
go test ./lessons/10-interfaces/solutions/... -v 2>&1 | tail -30
make test
golangci-lint run ./...
go vet ./...
```

Expected: gofmt empty; exercises pass vacuously (TestGreeters + TestGreeterSlice + TestJSONStore + TestMemoryStore — all empty-skeleton); solutions pass all sub-tests (6+1 in greet, 2 test functions × 3 contract checks in store ≈ 6 assertions); make test green; lint 0 issues; vet clean.

- [ ] **Step 12:** Solutions binary smoke test (JSONStore):

```bash
TMP=$(mktemp /tmp/expenses-XXXXXX.json)
rm $TMP
go run ./lessons/10-interfaces/solutions/cmd/expenses -store=json -file=$TMP add 2026-05-20 4.50 coffee
go run ./lessons/10-interfaces/solutions/cmd/expenses -store=json -file=$TMP add 2026-05-20 12.00 lunch
go run ./lessons/10-interfaces/solutions/cmd/expenses -store=json -file=$TMP list
go run ./lessons/10-interfaces/solutions/cmd/expenses -store=json -file=$TMP summary
rm $TMP
```

Expected: 2 add lines, 2 list lines, summary with 2-line per-category total. No bar chart (lesson 10's cmdSummary omits it).

- [ ] **Step 13:** Solutions binary smoke test (MemoryStore — single-invocation only, since memory doesn't persist):

```bash
go run ./lessons/10-interfaces/solutions/cmd/expenses -store=mem add 2026-05-20 4.50 coffee
```

Expected: `added: 2026-05-20  €4.50    coffee` printed, then process exits. (A follow-up `... -store=mem list` returns nothing — by design.)

- [ ] **Step 14:** Commit:

```bash
git add lessons/10-interfaces/exercises/expense/ lessons/10-interfaces/exercises/store/ lessons/10-interfaces/exercises/cmd/ \
        lessons/10-interfaces/solutions/expense/ lessons/10-interfaces/solutions/store/ lessons/10-interfaces/solutions/cmd/
git commit -m "feat(lesson-10): main — Store interface + JSONStore + MemoryStore + refactored CLI"
```

---

## Task 4: Author the slide deck

Replace `lessons/10-interfaces/slides/slides.md`. 4 concepts: implicit satisfaction → small interfaces → `any` + type assertions → interface design. The controller writes this directly inline (Plan K/L pattern — embedded literal markdown skipped from this plan; structure verification below).

**Files:**
- Replace: `lessons/10-interfaces/slides/slides.md`

**Structure to produce:**

```markdown
<div class="title-slide-grid">
  ... (lesson-number 10, Phase 2, h1: Interfaces, learning goal)
</div>

---

## What we'll cover
- ... 4 bullets matching the 4 concepts

---

## Concept 1: Implicit satisfaction
### Motivation
... why interfaces matter in Go vs other languages — no `implements`, no nominal typing
### The basics
... fmt.Stringer as the canonical example; English/Finnish as the warmup tie-in
### A worked example
... a function that takes a Greeter, called with English then Finnish
### Common mistake
... declaring a type "implements" an interface (impossible in Go)
### Recap

---

## Concept 2: Small interfaces (`io.Reader`/`io.Writer`/`error`)
### Motivation
... interface segregation: small interfaces compose; big interfaces don't
### The basics
... showing io.Reader, io.Writer, error — all 1-method interfaces
### A worked example
... the Store interface from this lesson (2 methods is still small)
### Common mistake
... designing a "Service" interface with 20 methods nobody fully implements
### Recap

---

## Concept 3: `any` and type assertions
### Motivation
... `any` = `interface{}` (alias since Go 1.18); when to use it; when not to
### The basics
... v, ok := x.(T); the panic form vs the comma-ok form
### A worked example
... a tiny "stringer" helper that handles fmt.Stringer specially
### Common mistake
... using any everywhere instead of writing focused interfaces
### Recap

---

## Concept 4: Interface design
### Motivation
... "accept interfaces, return structs" — the Go community rule
### The basics
... return concrete types from constructors; accept interfaces in functions
### A worked example
... NewJSONStore returns *JSONStore (concrete); cmd/expenses takes store.Store (interface)
### Common mistake
... taking *Interface as a parameter (pointer-to-interface is almost always wrong)
### Recap

---

## Practice
### Warm-up
... point to exercises/warmup/greet/
### Main
... point to exercises/{expense, store, cmd/expenses}/
### Note
... bar-chart deferred; -store=mem doesn't persist; etc.

---

## What we learned
... 5-bullet recap

---

## Up next
Lesson 11 — Errors. ...
```

**Structure check after writing:**

```bash
grep -c '<div class="title-slide-grid">' lessons/10-interfaces/slides/slides.md  # 1
grep -c "<h1>Interfaces</h1>" lessons/10-interfaces/slides/slides.md             # 1
grep -c "^## Concept " lessons/10-interfaces/slides/slides.md                     # 4
grep -c "^### Motivation$" lessons/10-interfaces/slides/slides.md                 # 4
grep -c "^### Common mistake$" lessons/10-interfaces/slides/slides.md             # 4
grep -c "^## What we learned" lessons/10-interfaces/slides/slides.md              # 1
grep -c "^## Up next" lessons/10-interfaces/slides/slides.md                       # 1
```

Commit:

```bash
git add lessons/10-interfaces/slides/slides.md
git commit -m "feat(lesson-10): slides — Interfaces (4 concepts)"
```

---

## Task 5: Author the README

Sections: Learning goals, Prerequisites, What's different (the `-store=mem` warning + the Store interface introduction), Concepts (mirroring the 4 slide concepts with Common-mistake examples), Exercise: warm-up, Exercise: main, How to run (incl. both -store flavours), Going further (Read + Try).

**Files:**
- Replace: `lessons/10-interfaces/README.md`

**Structure check:**

```bash
for h in "^# Lesson 10: Interfaces$" "^## Learning goals$" "^## Prerequisites$" "^## What's different about this lesson$" "^## Concepts$" "^## Exercise: warm-up$" "^## Exercise: main$" "^## How to run$" "^## Going further$" "^### Read$" "^### Try$"; do
  echo "  $h => $(grep -c "$h" lessons/10-interfaces/README.md)"
done
grep -c '^\*\*Common mistake\.\*\*' lessons/10-interfaces/README.md  # 4
grep -c "gofmt" lessons/10-interfaces/README.md                       # >= 1
grep -c "go vet" lessons/10-interfaces/README.md                      # >= 1
```

Commit:

```bash
git add lessons/10-interfaces/README.md
git commit -m "docs(lesson-10): README — Interfaces self-study"
```

---

## Task 6: End-to-end verification

- [ ] **Step 1:** `make test` — all packages green (lessons 01-10 + tools).

- [ ] **Step 2:** `make test-exercises` — lesson 10's tests pass vacuously.

- [ ] **Step 3:** `make test-lesson LESSON=10-interfaces` — solutions pass fully (TestGreeters 6, TestGreeterSlice 1, TestJSONStore + TestMemoryStore via runStoreContract — each runs 3 contract checks).

- [ ] **Step 4:** Solutions binary smoke tests (both -store kinds; same as Task 3 Steps 12-13).

- [ ] **Step 5:** `golangci-lint run ./...` — 0 issues.

- [ ] **Step 6:** `go vet ./...` — clean.

- [ ] **Step 7:** `make slides-build` — `dist/index.html` lists lesson 10.

- [ ] **Step 8:** Final sanity:

```bash
git status
make test
git log --no-show-signature --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 6 commits on the branch.

No commit for this task.

---

## Done definition

After Task 6:

- `lessons/10-interfaces/` contains 16 files in the subpackage layout.
- `make test` passes; `make test-exercises` shows lesson 10's tests passing vacuously.
- Solutions binary works for both `-store=json` and `-store=mem`.
- `make slides-build` produces `dist/` containing the lesson.
- `golangci-lint run ./...` reports 0 issues.
- `go vet ./...` is clean.
- 6 commits on the branch.

## What ships next

**Plan N — Lesson 11 (Errors).** Same per-lesson pattern. Lesson 11 formalises error handling: sentinel errors (`var ErrNotFound = errors.New(...)`), wrapping with `%w`, `errors.Is`/`As`, custom error types. The tracker gets richer storage errors (path + line number context); the README explains why the Phase 1 `if err != nil { return err }` pattern needed an upgrade. Tracker continues evolving.
