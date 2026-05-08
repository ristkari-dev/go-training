# Plan D — Lesson 01 (Hello, Go) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author the first lesson of Phase 1 — students go from "Go is installed" to "I have written and run my first Go function with formatted output." Six tasks: scaffold the lesson skeleton, author the warm-up exercise, author the main exercise, author the slide deck, author the self-study README, and end-to-end verify.

**Architecture:** The platform from Plan C produces a 12-file lesson skeleton via `make new-lesson NAME=01-hello`. Plan D replaces every `TODO` placeholder in that skeleton with concrete lesson content: two pure functions (`WarmupHello`, `WarmupGreet`) for the warm-up; a third pure function (`FormatExpense`) for the main exercise; a heavy-explanatory slide deck covering four concepts (`go run`, modules, `package main` + `func main()`, `fmt` printing); a self-study README mirroring the deck. All exercises are tested via `go test`; no live `main()` is part of the lesson tree (students experiment with `go run` outside the lesson folder).

**Tech Stack:** Go 1.23 stdlib only (`fmt`, `testing`). No third-party dependencies. Reveal.js 5.1.0 for the deck (vendored from Plan A).

---

## Scope

Plan D produces lesson 01 only. After Plan D lands:

- `lessons/01-hello/` is a complete, teachable lesson with all four parts populated (README, slides, exercises, solutions).
- `make test` passes (the solutions' tests are green).
- `make test-exercises` shows the lesson's exercise tests failing by design (panic stubs).
- `make slides-dev LESSON=01-hello` serves the deck on `localhost:8000`.
- The lesson is included in `dist/index.html` produced by `make slides-build`.

**Out of scope (handled by future plans):**
- Lesson 02 content (Plan E): variables, types, operators.
- Lessons 03-08 (Plans F-K).
- Any platform changes — Plan C already shipped them.

---

## File Structure

After Plan D:

```
lessons/
├── 01-hello/                              (new — Plan D's deliverable)
│   ├── README.md                          (Task 5 — full lesson prose)
│   ├── slides/
│   │   ├── index.html                     (Task 1 — scaffolded; unchanged)
│   │   ├── slides.md                      (Task 4 — full lesson content)
│   │   └── assets/.gitkeep                (Task 1 — scaffolded; unchanged)
│   ├── exercises/
│   │   ├── warmup.go                      (Task 2 — WarmupHello + WarmupGreet stubs)
│   │   ├── warmup_test.go                 (Task 2 — failing tests)
│   │   ├── main.go                        (Task 3 — Greet + FormatExpense stub)
│   │   └── main_test.go                   (Task 3 — failing tests)
│   └── solutions/
│       ├── warmup.go                      (Task 2 — implementations)
│       ├── warmup_test.go                 (Task 2 — same tests)
│       ├── main.go                        (Task 3 — implementations)
│       └── main_test.go                   (Task 3 — same tests)
└── .gitkeep                               (REMOVED in Task 1)
```

### Decomposition rationale

- **Task 1** scaffolds via `make new-lesson` so the scaffolder is genuinely exercised on a real lesson name. The scaffolded placeholder content is committed as a clear "scaffolded skeleton" point so subsequent commits show meaningful diffs of placeholder → real content.
- **Tasks 2 and 3** are split (warm-up vs main) because they're conceptually different (warm-up = pure string functions; main = `fmt.Sprintf` formatting). Separate commits make review easier.
- **Tasks 4 and 5** are split (slides vs README) because each is a substantial single-file authoring task. Slides land first because the README references the slide structure conceptually.
- **Task 6** is verification with no commit — confirms the lesson works end-to-end.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/` for every command. Do not navigate above this.
- **Branch:** `feature/plan-d-lesson-01` (you'll create this in Task 1).
- **Commit messages:** Conventional Commits (`feat:`, `chore:`, `docs:`).
- **Verification at end of every task:** `make test` passes, `git status` clean.

---

## Task 1: Scaffold lesson 01

Generate the lesson skeleton via `make new-lesson` and commit it as a clear baseline. Remove the now-redundant `lessons/.gitkeep`.

**Files:**
- Create: `lessons/01-hello/` (12 files via the scaffolder)
- Delete: `lessons/.gitkeep`

- [ ] **Step 1: Confirm you're on the feature branch with the plan committed**

The branch `feature/plan-d-lesson-01` already exists and the Plan D document is the only commit ahead of `main`. Verify:

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --oneline main..HEAD
```

Expected: `## feature/plan-d-lesson-01`, clean working tree, exactly one commit ahead of `main` ("Add Plan D: Lesson 01 (Hello, Go) implementation plan"). If you see something else, STOP and report — do not try to recreate the branch.

- [ ] **Step 2: Scaffold the lesson**

```bash
make new-lesson NAME=01-hello
```

Expected: prints `created lesson 01-hello under lessons/`. The scaffolder produces 12 files under `lessons/01-hello/` with template placeholder content (TODOs, `WarmupGreet`, `Greet`).

- [ ] **Step 3: Verify the 12-file tree exists**

```bash
find lessons/01-hello -type f | sort
```

Expected (12 lines):

```
lessons/01-hello/README.md
lessons/01-hello/exercises/main.go
lessons/01-hello/exercises/main_test.go
lessons/01-hello/exercises/warmup.go
lessons/01-hello/exercises/warmup_test.go
lessons/01-hello/slides/assets/.gitkeep
lessons/01-hello/slides/index.html
lessons/01-hello/slides/slides.md
lessons/01-hello/solutions/main.go
lessons/01-hello/solutions/main_test.go
lessons/01-hello/solutions/warmup.go
lessons/01-hello/solutions/warmup_test.go
```

- [ ] **Step 4: Remove `lessons/.gitkeep` (no longer needed)**

```bash
rm lessons/.gitkeep
test ! -f lessons/.gitkeep && echo OK
```

Expected: `OK`.

- [ ] **Step 5: Commit the scaffolded skeleton**

```bash
git add lessons/01-hello/ lessons/.gitkeep
git commit -m "feat(lessons): scaffold lesson 01-hello skeleton"
```

(The `git add lessons/.gitkeep` records the deletion. `git diff --cached --stat` should show the gitkeep as deleted and 12 new files.)

- [ ] **Step 6: Verify make test still passes**

```bash
make test
```

Expected: every package passes. `make test` excludes `*/exercises/*`, so the lesson's panic-stub tests don't break the build. The `lessons/01-hello/solutions/...` packages pass because the scaffolded solution implements `WarmupGreet` and `Greet` already.

- [ ] **Step 7: Sanity check git status**

```bash
git status
git log --oneline main..HEAD
```

Expected: clean working tree, one commit on the branch (the scaffold commit).

---

## Task 2: Author the warm-up exercise

Replace the four warm-up files (exercises + solutions, each with code + test) with lesson 01's actual warm-up: two pure functions, `WarmupHello` and `WarmupGreet`. The scaffolder shipped a single `WarmupGreet`; we extend to two functions and adjust the tests accordingly.

**Files:**
- Replace: `lessons/01-hello/exercises/warmup.go`
- Replace: `lessons/01-hello/exercises/warmup_test.go`
- Replace: `lessons/01-hello/solutions/warmup.go`
- Replace: `lessons/01-hello/solutions/warmup_test.go`

- [ ] **Step 1: Replace `lessons/01-hello/exercises/warmup.go`** with:

```go
// Package exercises is the starter code for lesson 01: Hello, Go.
//
// This file holds the WARM-UP exercise. Two small functions to build
// muscle memory for "a Go function that returns a string." Make the
// failing tests in warmup_test.go pass.
package exercises

// WarmupHello returns the greeting "Hello, Go!".
func WarmupHello() string {
	panic("TODO: return the string \"Hello, Go!\"")
}

// WarmupGreet returns "Hello, <name>!" for the given name.
//
// Examples:
//   WarmupGreet("World") → "Hello, World!"
//   WarmupGreet("Aki")   → "Hello, Aki!"
//
// Hint: Go strings concatenate with the + operator.
func WarmupGreet(name string) string {
	panic("TODO: return \"Hello, \" + name + \"!\"")
}
```

- [ ] **Step 2: Replace `lessons/01-hello/exercises/warmup_test.go`** with:

```go
package exercises

import "testing"

func TestWarmupHello(t *testing.T) {
	got := WarmupHello()
	want := "Hello, Go!"
	if got != want {
		t.Errorf("WarmupHello() = %q, want %q", got, want)
	}
}

func TestWarmupGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"World", "World", "Hello, World!"},
		{"Aki", "Aki", "Hello, Aki!"},
		{"empty", "", "Hello, !"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupGreet(tc.in)
			if got != tc.want {
				t.Errorf("WarmupGreet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/01-hello/solutions/warmup.go`** with:

```go
// Package solutions is the reference implementation for lesson 01: Hello, Go.
//
// This file holds the warm-up reference solution.
package solutions

// WarmupHello returns the greeting "Hello, Go!".
func WarmupHello() string {
	return "Hello, Go!"
}

// WarmupGreet returns "Hello, <name>!" for the given name.
func WarmupGreet(name string) string {
	return "Hello, " + name + "!"
}
```

- [ ] **Step 4: Replace `lessons/01-hello/solutions/warmup_test.go`** with the same content as `exercises/warmup_test.go`, but the `package` declaration is `package solutions`:

```go
package solutions

import "testing"

func TestWarmupHello(t *testing.T) {
	got := WarmupHello()
	want := "Hello, Go!"
	if got != want {
		t.Errorf("WarmupHello() = %q, want %q", got, want)
	}
}

func TestWarmupGreet(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"World", "World", "Hello, World!"},
		{"Aki", "Aki", "Hello, Aki!"},
		{"empty", "", "Hello, !"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupGreet(tc.in)
			if got != tc.want {
				t.Errorf("WarmupGreet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

### Note on the "read a name from input" spec wording

The Phase 1 design spec describes the warm-up as: *"Print 'Hello, Go!'; then read a name from input and print 'Hello, <name>!'."* The implementation here makes `WarmupGreet(name string) string` a pure function (name passed as a parameter, no stdin reading) so the test suite can assert on its output directly. Students see `fmt.Scanln` and the broader "read from stdin" pattern conceptually via the slides (Concept 3's worked example with hardcoded variables previews the shape), and Lesson 03 (Control flow) is where stdin reading lands as part of `fmt.Scanln` being formally introduced.

This is a deliberate testability trade-off. If you'd rather the warm-up actually read from stdin, that's a one-paragraph follow-up — but it complicates the test harness materially for lesson 01.

- [ ] **Step 5: Verify the warm-up exercise tests fail by design**

```bash
go test ./lessons/01-hello/exercises/... -run Warmup -v 2>&1 | tail -20
```

Expected: both `TestWarmupHello` and `TestWarmupGreet` (with all 3 sub-tests) fail with panic messages mentioning the TODOs.

- [ ] **Step 6: Verify the warm-up solution tests pass**

```bash
go test ./lessons/01-hello/solutions/... -run Warmup -v
```

Expected: both `TestWarmupHello` and `TestWarmupGreet` PASS (4 PASS lines: TestWarmupHello + 3 sub-tests).

- [ ] **Step 7: Verify `make test` still passes**

```bash
make test
```

Expected: every package passes (excluding exercises).

- [ ] **Step 8: Commit**

```bash
git add lessons/01-hello/exercises/warmup.go lessons/01-hello/exercises/warmup_test.go \
        lessons/01-hello/solutions/warmup.go lessons/01-hello/solutions/warmup_test.go
git commit -m "feat(lesson-01): warm-up — WarmupHello + WarmupGreet"
```

---

## Task 3: Author the main exercise

Replace the four main-exercise files with lesson 01's main exercise: `Greet` (provided as a complete, tiny example) plus `FormatExpense` (the actual exercise — uses `fmt.Sprintf`).

**Files:**
- Replace: `lessons/01-hello/exercises/main.go`
- Replace: `lessons/01-hello/exercises/main_test.go`
- Replace: `lessons/01-hello/solutions/main.go`
- Replace: `lessons/01-hello/solutions/main_test.go`

- [ ] **Step 1: Replace `lessons/01-hello/exercises/main.go`** with:

```go
// Package exercises is the starter code for lesson 01: Hello, Go.
//
// This file holds the MAIN exercise. Greet is already implemented as a
// tiny worked example; FormatExpense is yours to write. Make the failing
// tests in main_test.go pass.
package exercises

// Greet returns a greeting like "Hello, World!".
//
// Provided as a worked example of a function that takes a string,
// concatenates with + operators, and returns a new string.
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// FormatExpense returns a formatted line for one expense.
//
// The format is "date  €amount  category" where amount is printed
// with exactly two decimal places.
//
// Examples:
//   FormatExpense("2026-05-07", 4.50, "coffee")     → "2026-05-07  €4.50  coffee"
//   FormatExpense("2026-05-07", 23.5, "groceries")  → "2026-05-07  €23.50  groceries"
//   FormatExpense("2026-04-30", 1234.5, "rent")     → "2026-04-30  €1234.50  rent"
//
// Hint: import "fmt" and use fmt.Sprintf with the format verb %.2f
// to print the amount with two decimal places.
func FormatExpense(date string, amount float64, category string) string {
	panic("TODO: implement FormatExpense (see hint in the doc comment)")
}
```

- [ ] **Step 2: Replace `lessons/01-hello/exercises/main_test.go`** with:

```go
package exercises

import "testing"

func TestGreet(t *testing.T) {
	got := Greet("World")
	want := "Hello, World!"
	if got != want {
		t.Errorf("Greet(%q) = %q, want %q", "World", got, want)
	}
}

func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name     string
		date     string
		amount   float64
		category string
		want     string
	}{
		{"coffee", "2026-05-07", 4.50, "coffee", "2026-05-07  €4.50  coffee"},
		{"groceries", "2026-05-07", 23.5, "groceries", "2026-05-07  €23.50  groceries"},
		{"big-rent", "2026-04-30", 1234.5, "rent", "2026-04-30  €1234.50  rent"},
		{"zero", "2026-05-07", 0, "free", "2026-05-07  €0.00  free"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExpense(tc.date, tc.amount, tc.category)
			if got != tc.want {
				t.Errorf("FormatExpense(%q, %g, %q) = %q, want %q",
					tc.date, tc.amount, tc.category, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 3: Replace `lessons/01-hello/solutions/main.go`** with:

```go
// Package solutions is the reference implementation for lesson 01: Hello, Go.
package solutions

import "fmt"

// Greet returns a greeting like "Hello, World!".
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// FormatExpense returns a formatted line for one expense.
func FormatExpense(date string, amount float64, category string) string {
	return fmt.Sprintf("%s  €%.2f  %s", date, amount, category)
}
```

- [ ] **Step 4: Replace `lessons/01-hello/solutions/main_test.go`** with the same content as `exercises/main_test.go`, but `package solutions`:

```go
package solutions

import "testing"

func TestGreet(t *testing.T) {
	got := Greet("World")
	want := "Hello, World!"
	if got != want {
		t.Errorf("Greet(%q) = %q, want %q", "World", got, want)
	}
}

func TestFormatExpense(t *testing.T) {
	cases := []struct {
		name     string
		date     string
		amount   float64
		category string
		want     string
	}{
		{"coffee", "2026-05-07", 4.50, "coffee", "2026-05-07  €4.50  coffee"},
		{"groceries", "2026-05-07", 23.5, "groceries", "2026-05-07  €23.50  groceries"},
		{"big-rent", "2026-04-30", 1234.5, "rent", "2026-04-30  €1234.50  rent"},
		{"zero", "2026-05-07", 0, "free", "2026-05-07  €0.00  free"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatExpense(tc.date, tc.amount, tc.category)
			if got != tc.want {
				t.Errorf("FormatExpense(%q, %g, %q) = %q, want %q",
					tc.date, tc.amount, tc.category, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Verify the main exercise tests fail by design (FormatExpense panics; Greet passes since it's provided)**

```bash
go test ./lessons/01-hello/exercises/... -v 2>&1 | tail -25
```

Expected: `TestGreet` PASSes, `TestFormatExpense` (and all 4 sub-tests) FAILs with panic about implementing FormatExpense. `TestWarmupHello` and `TestWarmupGreet` (from Task 2) also still fail — that's by design.

- [ ] **Step 6: Verify the main solution tests pass**

```bash
go test ./lessons/01-hello/solutions/... -v
```

Expected: `TestGreet`, `TestFormatExpense` (4 sub-tests), `TestWarmupHello`, `TestWarmupGreet` (3 sub-tests) all PASS.

- [ ] **Step 7: Verify `make test` still passes**

```bash
make test
```

Expected: every package passes.

- [ ] **Step 8: Commit**

```bash
git add lessons/01-hello/exercises/main.go lessons/01-hello/exercises/main_test.go \
        lessons/01-hello/solutions/main.go lessons/01-hello/solutions/main_test.go
git commit -m "feat(lesson-01): main — Greet + FormatExpense"
```

---

## Task 4: Author the slide deck

Replace `lessons/01-hello/slides/slides.md` with the lesson 01 slide content. The deck follows the heavy-explanatory pattern: 4 concept blocks (each with Motivation → Basics → Worked example → Common mistake → Recap), framing slides, and a Practice slide pointing at the exercises.

**Files:**
- Replace: `lessons/01-hello/slides/slides.md`

- [ ] **Step 1: Replace `lessons/01-hello/slides/slides.md`** with the content below.

> CRITICAL: The block below is wrapped in four-backtick fences purely as a documentation device — the inner content has three-backtick code blocks. When you write the actual file, use only the three-backtick fences inside. The file should start with `## Lesson 01` (no leading fence).

````markdown
## Lesson 01

# Hello, Go

Learning goal: get a Go program running on your machine, understand `package main` and `func main()`, and use `fmt.Println` / `fmt.Printf` to print formatted output.

---

## What we'll cover

- Installing Go and verifying with `go version`.
- Running your first Go file with `go run`.
- Initialising a project with `go mod init` and what `go.mod` is.
- The role of `package main` and `func main()`.
- Printing with `fmt.Println` and `fmt.Printf`.

---

## Concept 1: From Go installed to running code

### Motivation

Before we write Go programs, we need to be able to run them. Unlike scripting languages, Go is compiled — but we don't have to deal with that explicitly for now. Go's tooling handles compilation behind the scenes.

For small programs you can think of `go run file.go` as "run this Go file" the same way you'd think `python file.py` runs a Python script. This first concept is about getting from "Go is installed" to "I just ran my first Go program."

---

### The basics

Save this as `hello.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

Then run:

```bash
go run hello.go
```

You should see:

```
Hello, Go!
```

That's it. You just compiled and ran a Go program.

---

### A worked example

Let's start the running theme of this course: a personal expense tracker. By the end of Phase 1, you'll have built a small CLI for it. For now, just print a hello-world version:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello! You spent €23.50 on coffee today.")
}
```

Run it:

```bash
go run hello.go
```

You'll see:

```
Hello! You spent €23.50 on coffee today.
```

That `€23.50` is hardcoded as a string for now. Lesson 02 introduces variables and types so the `23.50` can be a number you compute with.

---

### Common mistake

A classic mistake: forgetting `package main`.

```go
import "fmt"

func main() {
	fmt.Println("Hello!")
}
```

Run it and Go complains:

```
hello.go:1:1: expected 'package', found 'import'
```

Every Go file starts with a `package` declaration. For executable programs, that line is exactly `package main`.

---

### Recap

- `go run file.go` compiles and runs a Go file in one shot.
- Every Go file starts with `package <something>`.
- An executable program lives in `package main`.

---

## Concept 2: Modules and the project structure

### Motivation

`go run hello.go` works for a single file. But real Go programs are organised into modules — a directory with a `go.mod` file at the top. The module is what `go build`, `go test`, and `go run ./...` understand as "this project."

For Phase 1, every lesson lives inside the course's single module. You'll create modules of your own when you start projects from scratch.

---

### The basics

Create a new directory and initialise a module:

```bash
mkdir hello-project
cd hello-project
go mod init example.com/hello
```

You'll get a `go.mod` file:

```
module example.com/hello

go 1.23
```

The `module` line is the *import path*. The `go` line is the minimum Go version. That's all you need to start.

---

### A worked example

Add a `main.go` next to the `go.mod`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello from a real module!")
}
```

Now you can run:

```bash
go run .
```

The `.` means "the current package" — Go looks at all `.go` files in this directory, compiles them, and runs `main()`. That's the standard way to run programs from a module.

You can also still use `go run main.go` if you prefer.

---

### Common mistake

Confusing the module path with the directory name. The module path (`example.com/hello`) is just an identifier — it tells Go's tooling how to refer to your module if it were imported by someone else. It does not have to match your directory name.

The convention is to use a path you control on the internet — `github.com/yourusername/yourproject` is the most common form. A bare `my-project` works for personal use, but published modules need a real path.

---

### Recap

- A *module* is a directory with a `go.mod` file at the top.
- `go mod init <path>` creates the file.
- `go run .` runs all `.go` files in the current package.
- The module path is a logical identifier, not a directory name.

---

## Concept 3: The main package and main function

### Motivation

Every executable Go program has the same starting point: `func main()` inside `package main`. That's how Go knows which function to call when you run the binary.

This sounds simple, but the *combination* — special package name, special function name — is unique to Go and worth understanding.

---

### The basics

```go
package main

func main() {
	// This runs when the program starts.
}
```

Two rules for executable programs:

1. The package must be named `main`.
2. There must be exactly one function called `main` with no parameters and no return values.

If either is missing, `go run` complains.

---

### A worked example

Let's add some variables (technically lesson 02 territory, but useful here as a preview):

```go
package main

import "fmt"

func main() {
	date := "2026-05-07"
	amount := 4.50
	category := "coffee"
	fmt.Println(date, amount, category)
}
```

Output:

```
2026-05-07 4.5 coffee
```

`Println` separates its arguments with spaces and adds a newline at the end. Notice that `4.50` printed as `4.5` — Go drops trailing zeros for floats by default. We'll fix that with `Printf` in the next concept.

---

### Common mistake

Capitalisation matters. Go is case-sensitive everywhere:

```go
package main

func Main() {  // wrong: capital M
	fmt.Println("Hello!")
}
```

Run it and Go complains:

```
runtime.main_main·f: function main is undeclared in the main package
```

The fix is one keystroke: `Main` → `main`.

---

### Recap

- Executable programs need `package main` and `func main()`.
- `func main()` takes no arguments and returns nothing.
- Go is case-sensitive: `main` is not the same as `Main`.

---

## Concept 4: fmt and basic printing

### Motivation

You'll print things constantly while learning a language: to inspect values, to give feedback, to produce output. The `fmt` package is Go's primary tool for that.

`Println` is the simple "print this with a newline" function. `Printf` is the formatted version — it lets you control how numbers, strings, and other values look in the output. For lesson 01 we focus on these two, plus `Sprintf` (the same as `Printf` but returns a string instead of printing it).

---

### The basics

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")              // prints: Hello, world!
	fmt.Println("two", "things")              // prints: two things
	fmt.Printf("amount = %.2f\n", 4.5)        // prints: amount = 4.50
	fmt.Printf("%s ate %d fish\n", "Aki", 3)  // prints: Aki ate 3 fish
}
```

Three things to notice:

- `Println` joins its arguments with spaces and adds a newline.
- `Printf` uses *format verbs* (the `%` placeholders) but does NOT add a newline — you have to write `\n` yourself.
- `%s` is for strings, `%d` is for integers, `%f` is for floats. `%.2f` is "float with 2 decimal places."

---

### A worked example

The expense-tracker theme — print three expenses with two-decimal amounts:

```go
package main

import "fmt"

func main() {
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 4.50, "coffee")
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 23.50, "groceries")
	fmt.Printf("%s  €%.2f  %s\n", "2026-05-06", 1234.50, "rent")
}
```

Output:

```
2026-05-07  €4.50  coffee
2026-05-07  €23.50  groceries
2026-05-06  €1234.50  rent
```

Every amount is printed with exactly two decimal places, including `4.50` (which would have lost the trailing zero with `Println`). The `%.2f` verb gives consistent formatting.

---

### Common mistake

Forgetting the `\n` at the end of a `Printf`:

```go
fmt.Printf("Hello!")
fmt.Printf("World!")
```

Output:

```
Hello!World!
```

No newlines, no separation. `Println` adds the newline; `Printf` does not. If you want the line to end, include `\n` in the format string.

---

### Recap

- `fmt.Println` for simple "print + newline" output.
- `fmt.Printf` for formatted output. You write the newline yourself.
- `fmt.Sprintf` returns a formatted string instead of printing it.
- Format verbs: `%s` strings, `%d` ints, `%f` floats. `%.2f` for two decimals.

---

## Practice

### Warm-up

Two small functions in `exercises/warmup.go`:

- `WarmupHello()` returns `"Hello, Go!"`.
- `WarmupGreet(name)` returns `"Hello, <name>!"` for any name.

Both are one line of code using string concatenation. Make the failing tests pass.

```bash
cd lessons/01-hello/exercises
go test -run Warmup -v
```

---

### Main

`FormatExpense(date, amount, category)` in `exercises/main.go` returns a column-formatted string like `"2026-05-07  €4.50  coffee"`. Use `fmt.Sprintf` with `%s` and `%.2f` verbs.

The starter doesn't import `fmt` — you'll add `import "fmt"` at the top of `main.go`.

```bash
cd lessons/01-hello/exercises
go test -v
```

Note:
For live: walk through `go run hello.go` on the projector. Hit the common mistakes deliberately so students see the error messages — `package main` missing, capital `Main`, missing `\n`. Skip the longer worked example in concept 2 if running short on time; the exercises cover that ground.

---

## What we learned

- `go run file.go` compiles and runs a Go program in one step.
- A *module* is a directory with `go.mod`; `go mod init` creates it.
- Executable programs are `package main` with `func main()`.
- `fmt.Println` for simple printing; `fmt.Printf` for formatted output; `fmt.Sprintf` returns the formatted string.
- Go is case-sensitive — `main` not `Main`.

---

## Up next

Lesson 02 — Variables, types, operators.
````

- [ ] **Step 2: Verify the deck renders correctly**

```bash
make slides-dev LESSON=01-hello &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/01-hello/slides/slides.md
curl -sS http://localhost:8000/lessons/01-hello/slides/slides.md | grep -c "^## Concept "
curl -sS http://localhost:8000/lessons/01-hello/slides/slides.md | grep -c "^### Motivation$"
curl -sS http://localhost:8000/lessons/01-hello/slides/slides.md | grep -c "^### Common mistake$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, then `4`, `4`, `4` (four concept blocks each with Motivation and Common mistake subsections).

- [ ] **Step 3: Verify markdown structure with a few greps**

```bash
grep -c "^## Lesson 01" lessons/01-hello/slides/slides.md
grep -c "^# Hello, Go" lessons/01-hello/slides/slides.md
grep -c "^## What we learned" lessons/01-hello/slides/slides.md
grep -c "^## Up next" lessons/01-hello/slides/slides.md
```

Expected: `1`, `1`, `1`, `1`.

- [ ] **Step 4: Commit**

```bash
git add lessons/01-hello/slides/slides.md
git commit -m "feat(lesson-01): slides — Hello, Go (4 concepts)"
```

---

## Task 5: Author the README

Replace `lessons/01-hello/README.md` with the full lesson 01 self-study prose. Mirrors the slide narrative, fills in concrete prerequisites, gives explicit "How to run" commands, and provides Going further links.

**Files:**
- Replace: `lessons/01-hello/README.md`

- [ ] **Step 1: Replace `lessons/01-hello/README.md`** with the markdown below.

> Same convention as Task 4: the four-backtick wrapper is a documentation device. In the actual file, use only three-backtick fences. The file starts with `# Lesson 01: Hello, Go`.

````markdown
# Lesson 01: Hello, Go

## Learning goals

- Verify your Go installation and run a Go program from the command line.
- Initialise a new Go module with `go mod init` and understand what `go.mod` contains.
- Recognise the role of `package main` and `func main()` in executable Go programs.
- Print formatted output with `fmt.Println` and `fmt.Printf`, including float formatting with `%.2f`.

## Prerequisites

This is the first lesson — no prior lessons assumed. You should have:

- Go 1.23 or newer installed (`go version` should print something like `go version go1.23.0 darwin/arm64`).
- A terminal you're comfortable typing in.
- A text editor with Go support — VS Code, GoLand, vim, emacs all work.

If you don't have Go installed yet, the official guide is at <https://go.dev/doc/install>.

## Concepts

### From Go installed to running code

Go is a compiled language, but for everyday work you don't need to think about that explicitly. The `go run` command compiles your file in the background and runs the result. Save a file as `hello.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

Then run it:

```bash
go run hello.go
```

You'll see `Hello, Go!` printed. That's the smallest complete Go program — three required pieces: the package declaration on line 1, the `import "fmt"` to bring in the printing package, and `func main()` as the program's entry point.

### Modules and the project structure

A single `.go` file works for tiny demos, but real Go projects are organised as *modules*. A module is a directory with a `go.mod` file at the top. You create one with:

```bash
go mod init example.com/hello
```

The `go.mod` file looks like this:

```
module example.com/hello

go 1.23
```

`example.com/hello` is the module path — a logical identifier that other modules would use to import yours. `go 1.23` declares the minimum Go version. Inside a module, you can run `go run .` (the dot means "the current package") and Go figures out which files to compile. This is the standard way to run programs once a project has more than one file.

### The main package and main function

Every executable Go program has the same entry point: `func main()` inside `package main`. The package name `main` is what tells Go this is a runnable program (rather than a library). The `main` function takes no parameters and returns nothing — its only job is to run when the program starts.

If you write `package mainz` or `func Main()` (capital M), Go won't run the program. It'll either complain at compile time or report that no entry point exists. Case matters.

### Formatting output with fmt

The `fmt` package handles printing. Two functions cover most of what you need in lesson 01:

- `fmt.Println(args...)` — prints its arguments separated by spaces and adds a newline. Best for quick output.
- `fmt.Printf(format, args...)` — uses *format verbs* like `%s`, `%d`, `%f` to control how each argument is printed. You write the newline yourself with `\n`.

The format verbs you'll use most:

- `%s` — a string
- `%d` — a decimal integer
- `%f` — a floating-point number (default formatting, may have many decimals)
- `%.2f` — a float with exactly two decimal places (great for currency)

Example:

```go
fmt.Printf("%s  €%.2f  %s\n", "2026-05-07", 4.50, "coffee")
```

prints:

```
2026-05-07  €4.50  coffee
```

`fmt.Sprintf` is the same as `Printf` but returns the formatted string instead of printing it. You'll use `Sprintf` in this lesson's main exercise.

## Exercise: warm-up

Open `exercises/warmup.go`. There are two small functions to implement:

- `WarmupHello() string` — return the literal string `"Hello, Go!"`.
- `WarmupGreet(name string) string` — return `"Hello, <name>!"` where `<name>` is the input.

Both are one-line functions using string concatenation (`+`). Make the failing tests in `warmup_test.go` pass.

## Exercise: main

Open `exercises/main.go`. The `Greet` function is already provided as a worked example; your job is `FormatExpense`.

`FormatExpense(date string, amount float64, category string) string` returns a column-formatted line like `"2026-05-07  €4.50  coffee"`. Use `fmt.Sprintf` with the `%s` and `%.2f` format verbs to build the string.

You'll need to add `import "fmt"` at the top of the file — the starter doesn't import it because the panic stub doesn't need it.

The tests cover several cases including amounts with zero decimals (must still print as `.00`), big numbers (no comma separators), and small numbers.

## How to run

```bash
cd lessons/01-hello/exercises
go test -run Warmup -v   # warm-up only
go test -v                # everything
```

Once both exercises pass, take a look at `solutions/` to compare your code with the reference implementation. The solutions might use slightly different idioms — that's fine.

To experiment with `go run` on your own, create a `hello.go` file outside this folder (e.g., in `~/scratch/`) and try the examples from the slides. The lesson folder itself uses test-driven exercises, not runnable `main()` programs.

## Going further

### Read

- [A Tour of Go — Welcome](https://go.dev/tour/welcome/1) — the official interactive intro. Walks through similar material with playgrounds you can edit live.
- [The fmt package documentation](https://pkg.go.dev/fmt) — the canonical reference for format verbs. Skim the "Printing" section; you'll use it a lot.
- [Effective Go — Names](https://go.dev/doc/effective_go#names) — short read about Go's naming conventions. Worth reading early to internalise the style.

### Try

- **Currency localisation.** Modify your `FormatExpense` to take a currency symbol parameter (e.g. `"€"`, `"$"`, `"£"`) and use it in the output. No reference solution provided — pick the function signature that feels most readable to you.
- **Right-alignment.** Format verbs support width specifiers: `%8.2f` pads the float to width 8, right-aligned. Try printing a small list of expenses where every amount is right-aligned to width 8. The output should look like a clean column.
````

- [ ] **Step 2: Verify the README has the expected structure**

```bash
grep -c "^# Lesson 01: Hello, Go$" lessons/01-hello/README.md
grep -c "^## Learning goals$" lessons/01-hello/README.md
grep -c "^## Prerequisites$" lessons/01-hello/README.md
grep -c "^## Concepts$" lessons/01-hello/README.md
grep -c "^## Exercise: warm-up$" lessons/01-hello/README.md
grep -c "^## Exercise: main$" lessons/01-hello/README.md
grep -c "^## How to run$" lessons/01-hello/README.md
grep -c "^## Going further$" lessons/01-hello/README.md
grep -c "^### Read$" lessons/01-hello/README.md
grep -c "^### Try$" lessons/01-hello/README.md
```

Expected: every count is `1`.

- [ ] **Step 3: Commit**

```bash
git add lessons/01-hello/README.md
git commit -m "docs(lesson-01): README — Hello, Go self-study"
```

---

## Task 6: End-to-end verification

Confirm the lesson is teachable end to end: scaffolder still works, exercise tests fail by design, solution tests pass, slides render, README has the right structure, lint clean.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises). The lesson's solutions pass, the tools pass, no failures.

- [ ] **Step 2: Run exercise tests — must fail by design**

```bash
make test-exercises 2>&1 | tail -30
```

Expected: `TestWarmupHello`, `TestWarmupGreet`, `TestFormatExpense` all fail with panic messages mentioning their respective TODOs. `TestGreet` passes (it's provided as a worked example). Make exits 0 because the recipe uses `-` to ignore the failure.

- [ ] **Step 3: Run lesson-specific tests via `make test-lesson`**

```bash
make test-lesson LESSON=01-hello 2>&1 | tail -30
```

Expected: exercise tests fail (ignored), solution tests pass.

- [ ] **Step 4: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 5: Slides server smoke test**

```bash
make slides-dev LESSON=01-hello &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/01-hello/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/01-hello/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected:
- Line 1 (root): `302` (redirect to deck) or `200` if the root path resolves to the deck directly.
- Lines 2-4: `200` (or `301` for `index.html` due to nginx canonicalisation; either is acceptable).

- [ ] **Step 6: Build the static site and verify the lesson is listed**

```bash
make slides-build
test -f dist/index.html && echo OK
test -f dist/lessons/01-hello/slides/slides.md && echo OK
grep -q "Hello, Go" dist/index.html && echo "index lists lesson 01"
rm -rf dist
```

Expected: three `OK`/match lines.

- [ ] **Step 7: Final repository sanity check**

```bash
git status
make test
git log --oneline main..HEAD
```

Expected: clean tree, `make test` passes, 5 commits on the branch (scaffold + warm-up + main + slides + README).

This task makes no commit — it is verification only.

---

## Done definition

After Task 6, all of these are true:

- `lessons/01-hello/` contains 12 files (4 exercises, 4 solutions, slides + index + assets, README).
- `make test` passes; `make test-exercises` shows the lesson's exercise tests failing by design.
- `make test-lesson LESSON=01-hello` runs both exercise (fail) and solution (pass) tests.
- `make slides-dev LESSON=01-hello` serves the deck on `localhost:8000`.
- `make slides-build` produces a `dist/` containing the lesson's deck and the lesson title appears in the landing page.
- `golangci-lint run ./...` reports 0 issues.
- The git history is a clean sequence of 5 small, conventional commits.

## What ships next (after Plan D merges)

- **Plan E — Lesson 02: Variables, types, operators.** Same pattern: scaffold, warm-up, main, slides, README, verify. The lesson 02 main exercise will introduce typed expense variables and arithmetic over them, foreshadowing more of the capstone.
