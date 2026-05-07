# Plan C — Phase 1 Platform Updates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the lesson scaffolder and contributor docs to support Phase 1's warm-up + main exercise pattern and heavy-explanatory slide style — so subsequent per-lesson plans (D, E, F…) can run `make new-lesson` and start from a Phase-1-ready skeleton.

**Architecture:** Five small, focused commits to `tools/new-lesson/template/`, the scaffolder tests, and `CONTRIBUTING.md`. No code changes to the scaffolder Go program itself — its `WalkDir`-based template traversal already picks up any new `.tmpl` files. The `make new-lesson` interface stays identical; only the *output* changes.

**Tech Stack:** Go 1.23 (tests via `go test`), markdown for templates and docs, GNU Make for verification.

---

## Scope

Plan C is **platform-only**. It produces no lesson content. After Plan C lands:

- `make new-lesson NAME=NN-name` produces a lesson skeleton with:
  - `exercises/warmup.go` + `exercises/warmup_test.go` (in addition to `main.go` + `main_test.go`)
  - `solutions/warmup.go` + `solutions/warmup_test.go` (in addition to `main.go` + `main_test.go`)
  - A `slides/slides.md` pre-stocked with the heavy-explanatory pattern (3 concept blocks, each with Motivation/Basics/Worked example/Common mistake/Recap subsections, plus framing slides)
  - A `README.md` with separate "Exercise: warm-up" and "Exercise: main" sections
- `CONTRIBUTING.md` documents Phase 1 conventions (the `Warmup*` naming convention, the heavy-explanatory slide pattern, the lesson-4 testing-handover rule, the stdlib-only constraint).

**Out of scope (handled by future per-lesson plans):**

- Lesson 1 content (Plan D)
- Lesson 2 content (Plan E)
- Lessons 3-7 (Plans F-J)
- Lesson 8 capstone including the provided `storage` package (Plan K)

The naming convention `Plan D = lesson 1`, `Plan E = lesson 2`, etc. is mnemonic. If preferred, "Plan C-2", "Plan C-3" works equivalently — naming is decided when each per-lesson plan is written.

---

## File Structure

After Plan C:

```
tools/new-lesson/template/
├── README.md.tmpl                         (modified — task 3)
├── slides/
│   ├── index.html.tmpl                    (unchanged)
│   ├── slides.md.tmpl                     (rewritten — task 2)
│   └── assets/.gitkeep                    (unchanged)
├── exercises/
│   ├── warmup.go.tmpl                     (NEW — task 1)
│   ├── warmup_test.go.tmpl                (NEW — task 1)
│   ├── main.go.tmpl                       (unchanged)
│   └── main_test.go.tmpl                  (unchanged)
└── solutions/
    ├── warmup.go.tmpl                     (NEW — task 1)
    ├── warmup_test.go.tmpl                (NEW — task 1)
    ├── main.go.tmpl                       (unchanged)
    └── main_test.go.tmpl                  (unchanged)

tools/new-lesson/main_test.go              (modified — task 1)
CONTRIBUTING.md                            (modified — task 4)
```

The scaffolder Go code (`tools/new-lesson/main.go`) is **not** modified — it walks any `.tmpl` files it finds, so new ones are picked up automatically.

### Decomposition rationale

- **Task 1** bundles the new template files with the scaffolder-test update because the two are tightly coupled: the test's `want` list grows from 8 to 12 paths and would either fail without the files (if test updated first) or be misleading (if files added first without updating the test). One coherent commit.
- **Task 2** is the slides.md skeleton rewrite — a substantial markdown file authored once and used by every Phase 1 lesson. Its own commit so it can be reviewed in isolation.
- **Task 3** is the README.md.tmpl change — small but author-facing, separate commit for clarity.
- **Task 4** is documentation only.
- **Task 5** is verification with no commit — confirms `make new-lesson` produces the expected 12-file tree.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/` for every command. Do not navigate above this.
- **Branch:** `feature/plan-c-phase-1` (already created; spec commits already there).
- **Commit messages:** Conventional Commits (`feat:`, `chore:`, `docs:`, `test:`, `build:`).
- **Verification:** Each task ends with `go test ./tools/new-lesson/...` passing and `git status` clean.

---

## Task 1: Add warm-up template files + update scaffolder tests

The lesson template currently produces 8 files per scaffolded lesson. Add 4 more (warm-up exercise/solution + their tests), and update the scaffolder's `TestScaffoldCreatesExpectedTree` to expect the new layout.

The warm-up template ships with a `WarmupGreet` function — `Warmup`-prefixed to avoid colliding with the main exercise's `Greet`. This naming pattern is the convention every Phase 1 lesson follows.

**Files:**
- Create: `tools/new-lesson/template/exercises/warmup.go.tmpl`
- Create: `tools/new-lesson/template/exercises/warmup_test.go.tmpl`
- Create: `tools/new-lesson/template/solutions/warmup.go.tmpl`
- Create: `tools/new-lesson/template/solutions/warmup_test.go.tmpl`
- Modify: `tools/new-lesson/main_test.go`

- [ ] **Step 1: Update `TestScaffoldCreatesExpectedTree` to expect the new files (failing first)**

Open `tools/new-lesson/main_test.go` and find the `want` slice in `TestScaffoldCreatesExpectedTree`. Replace the current 8-entry slice with 12 entries:

```go
	want := []string{
		"01-hello/README.md",
		"01-hello/slides/index.html",
		"01-hello/slides/slides.md",
		"01-hello/slides/assets/.gitkeep",
		"01-hello/exercises/warmup.go",
		"01-hello/exercises/warmup_test.go",
		"01-hello/exercises/main.go",
		"01-hello/exercises/main_test.go",
		"01-hello/solutions/warmup.go",
		"01-hello/solutions/warmup_test.go",
		"01-hello/solutions/main.go",
		"01-hello/solutions/main_test.go",
	}
```

- [ ] **Step 2: Run the test — confirm it fails**

```bash
cd /Users/ristkari/code/private/go-training
go test ./tools/new-lesson/ -run TestScaffoldCreatesExpectedTree -v
```

Expected: `--- FAIL: TestScaffoldCreatesExpectedTree` with errors like `missing 01-hello/exercises/warmup.go` for each of the four new paths. The other tests (`TestParseName`, `TestScaffoldSubstitutesLessonInfo`, `TestScaffoldRefusesExistingDest`, `TestCLI*`) still pass.

- [ ] **Step 3: Create `tools/new-lesson/template/exercises/warmup.go.tmpl`**

```go
// Package exercises is the starter code for lesson {{.Number}}: {{.Title}}.
//
// This file holds the WARM-UP exercise. It is a small, focused exercise to
// build muscle memory for the lesson's key concept. Make the failing tests
// in warmup_test.go pass.
//
// The warm-up uses a Warmup* prefix on every exported name so it doesn't
// collide with identifiers in main.go. See CONTRIBUTING.md.
package exercises

// WarmupGreet returns a warm-up greeting like "Warm-up hello, World!".
func WarmupGreet(name string) string {
	panic("TODO: implement WarmupGreet")
}
```

- [ ] **Step 4: Create `tools/new-lesson/template/exercises/warmup_test.go.tmpl`**

```go
package exercises

import "testing"

func TestWarmupGreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"World", "Warm-up hello, World!"},
		{"Go", "Warm-up hello, Go!"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := WarmupGreet(tc.in)
			if got != tc.want {
				t.Errorf("WarmupGreet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 5: Create `tools/new-lesson/template/solutions/warmup.go.tmpl`**

```go
// Package solutions is the reference implementation for lesson {{.Number}}: {{.Title}}.
//
// This file holds the warm-up reference solution.
package solutions

// WarmupGreet returns a warm-up greeting.
func WarmupGreet(name string) string {
	return "Warm-up hello, " + name + "!"
}
```

- [ ] **Step 6: Create `tools/new-lesson/template/solutions/warmup_test.go.tmpl`**

```go
package solutions

import "testing"

func TestWarmupGreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"World", "Warm-up hello, World!"},
		{"Go", "Warm-up hello, Go!"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := WarmupGreet(tc.in)
			if got != tc.want {
				t.Errorf("WarmupGreet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 7: Run the scaffolder tests — confirm all pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: every test PASSes — `TestParseName` (3 sub-tests), `TestScaffoldCreatesExpectedTree`, `TestScaffoldSubstitutesLessonInfo`, `TestScaffoldRefusesExistingDest`, `TestCLIRejectsMissingName`, `TestCLIScaffoldsHappyPath`.

- [ ] **Step 8: Run lint**

```bash
golangci-lint run ./tools/new-lesson/...
```

Expected: 0 issues.

- [ ] **Step 9: Commit**

```bash
git add tools/new-lesson/template/ tools/new-lesson/main_test.go
git commit -m "feat(new-lesson): add warm-up template files alongside main"
```

---

## Task 2: Rewrite slides.md.tmpl with heavy-explanatory skeleton

The current `slides.md.tmpl` has a simple 5-slide skeleton (title, what we'll cover, a first example, recap, up next). Phase 1 needs the heavy-explanatory pattern: per-concept Motivation → Basics → Worked example → Common mistake → Recap. Pre-stocking three concept blocks gives lesson authors a meaningful starting point.

**Files:**
- Modify: `tools/new-lesson/template/slides/slides.md.tmpl`

- [ ] **Step 1: Replace the entire contents of `tools/new-lesson/template/slides/slides.md.tmpl`** with the following markdown content.

> The block below is wrapped in four-backtick fences purely as a documentation device (because the inner content has three-backtick code blocks). In the actual file, use only the three-backtick fences inside — do NOT include the outer four-backtick wrapper. The file should start with `## Lesson {{.Number}}` (no leading fence).

````markdown
## Lesson {{.Number}}

# {{.Title}}

Learning goal: TODO

---

## What we'll cover

- TODO
- TODO
- TODO

---

## Concept 1: TODO

### Motivation

TODO — why this concept exists, what problem it solves. One or two short paragraphs.

---

### The basics

```go
package main

import "fmt"

func main() {
	fmt.Println("TODO: minimal example introducing the concept")
}
```

---

### A worked example

TODO — a substantive example using the concept in a realistic context.

```go
package main

func main() {
	// TODO
}
```

---

### Common mistake

TODO — what NOT to do, with the bug the compiler/runtime would surface.

```go
// the wrong way
```

---

### Recap

- TODO
- TODO

---

## Concept 2: TODO

### Motivation

TODO

---

### The basics

```go
// minimal
```

---

### A worked example

TODO

```go
// concrete
```

---

### Common mistake

TODO

```go
// wrong way
```

---

### Recap

- TODO

---

## Concept 3: TODO

### Motivation

TODO

---

### The basics

```go
// minimal
```

---

### A worked example

TODO

```go
// concrete
```

---

### Common mistake

TODO

```go
// wrong way
```

---

### Recap

- TODO

---

## Practice

### Warm-up

TODO — short description of the warm-up exercise. Files: `exercises/warmup.go`, `exercises/warmup_test.go`.

```bash
cd lessons/{{.Name}}/exercises
go test -run Warmup -v
```

---

### Main

TODO — description of the main exercise. Files: `exercises/main.go`, `exercises/main_test.go`.

```bash
cd lessons/{{.Name}}/exercises
go test -v
```

Note:
Speaker notes for the live lecture go here. They render only in the speaker view (press `S`). Use `Note:` blocks throughout the deck to record what to actually say live; the slide prose is the self-study text.

---

## What we learned

- TODO
- TODO

---

## Up next

Lesson NN — TODO
````

> The `TODO` markers are intentional placeholders for the lesson author to replace per lesson. The triple-backtick code fences inside the markdown render as syntax-highlighted code blocks via reveal.js's highlight plugin.

- [ ] **Step 2: Verify the template still produces a parseable lesson**

Quick sanity check — scaffold a sandbox lesson, peek at the result, throw it away:

```bash
TMP=$(mktemp -d)
go run ./tools/new-lesson -name 99-demo -out "$TMP"
head -20 "$TMP/99-demo/slides/slides.md"
grep -c "^---$" "$TMP/99-demo/slides/slides.md"
rm -rf "$TMP"
```

Expected:
- `head -20` shows `## Lesson 99`, `# Demo`, `Learning goal: TODO`, etc.
- `grep -c "^---$"` shows roughly 25 separator lines (each `---` is a slide boundary).

- [ ] **Step 3: Confirm scaffolder tests still pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: every test PASSes (slides.md.tmpl content isn't checked by the existing tests; they only verify the file exists).

- [ ] **Step 4: Commit**

```bash
git add tools/new-lesson/template/slides/slides.md.tmpl
git commit -m "feat(new-lesson): rewrite slides.md skeleton for heavy-explanatory style"
```

---

## Task 3: Update README.md.tmpl with warm-up + main exercise sections

Currently the README template has one "Exercise" section. With the warm-up + main split, it needs two — and the "How to run" section needs to show both test commands.

**Files:**
- Modify: `tools/new-lesson/template/README.md.tmpl`

- [ ] **Step 1: Replace the entire contents of `tools/new-lesson/template/README.md.tmpl`** with the following markdown content.

> Same convention as Task 2: four-backtick wrapper is a documentation device only. In the actual file, use only the three-backtick fences inside. The file should start with `# Lesson {{.Number}}: {{.Title}}`.

````markdown
# Lesson {{.Number}}: {{.Title}}

## Learning goals

- TODO
- TODO
- TODO

## Prerequisites

- Lessons through {{.Number}} (this one is built on the prior lesson's material).

## Concepts

TODO — write the prose narrative that mirrors the slides for self-study readers. For Phase 1 lessons this should be several paragraphs per concept, not a recap.

## Exercise: warm-up

TODO — short description of the warm-up exercise (5-10 minutes). Open `exercises/warmup.go` and `exercises/warmup_test.go`. Make the failing tests pass.

## Exercise: main

TODO — description of the main exercise (30-45 minutes). Open `exercises/main.go` and `exercises/main_test.go`. Make the failing tests pass.

## How to run

```bash
cd lessons/{{.Name}}/exercises
go test -run Warmup -v   # warm-up only
go test -v                # everything
```

## Going further

### Read

- TODO — 1-3 short links (Go blog posts, std-lib docs, occasional book references).

### Try

- TODO — 1-2 stretch problems harder than the main exercise. Self-graded. There are no reference solutions in `solutions/` — the point is to push past the spec.
````

- [ ] **Step 2: Confirm `TestScaffoldSubstitutesLessonInfo` still passes**

That test reads the scaffolded README.md and checks for `"Lesson 05"`, `"Slices And Maps"`, and `"lessons/05-slices-and-maps/exercises"`. The new template still contains `# Lesson {{.Number}}: {{.Title}}` and `lessons/{{.Name}}/exercises`, so all three substrings should still appear after rendering.

```bash
go test ./tools/new-lesson/ -run TestScaffold -v
```

Expected: `TestScaffoldSubstitutesLessonInfo` PASSes alongside the other Scaffold tests.

- [ ] **Step 3: Commit**

```bash
git add tools/new-lesson/template/README.md.tmpl
git commit -m "feat(new-lesson): split README exercise section into warm-up + main"
```

---

## Task 4: Add Phase 1 conventions to CONTRIBUTING.md

Document the Phase 1 patterns the scaffolder now produces, so a future lesson author (or future-you, or another contributor) knows what they mean and why.

**Files:**
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Read the current `CONTRIBUTING.md`** to find the right insertion point. The file currently has these top-level sections (from Plan A):

- `# Contributing`
- `## Authoring a new lesson`
- `### The four-part structure`
- `### Slide style`
- `### Tests as the spec`
- `## Local workflow`
- `## Reveal.js`
- `## Commit messages`

Insert the new section **before** `## Local workflow` (after the existing `### Tests as the spec` content).

- [ ] **Step 2: Insert the new section**

Use the Edit tool to add this block immediately before the `## Local workflow` line. (Four-backtick wrapper is a documentation device — the inserted content begins with `## Phase 1 conventions (lessons 1-8)` and uses only three-backtick fences internally.)

````markdown
## Phase 1 conventions (lessons 1-8)

Phase 1 of the course (Foundations) introduces additional conventions on top of the four-part structure. The lesson scaffolder produces them by default; this section explains them.

### Warm-up + main exercise pattern

Every Phase 1 lesson has two exercises: a 5-10 minute warm-up that builds muscle memory for the lesson's key concept, and a 30-45 minute main exercise that applies it. They live alongside each other in the same Go package:

```
lessons/NN-name/exercises/
├── warmup.go         # exported names use a Warmup* prefix
├── warmup_test.go
├── main.go           # exported names are unprefixed
└── main_test.go
```

The `Warmup*` prefix on warm-up identifiers is mandatory. Two exercises in one package would otherwise collide on names like `Greet` or `Format`. The scaffolder's default warm-up exports `WarmupGreet` to model the pattern.

The same applies to `solutions/`.

**Lesson 7 onwards** uses subfolders instead of flat layout — that change is taught explicitly in the packages lesson, so it's not the default scaffold.

### Heavy-explanatory slide style

Phase 1 lessons use a textbook-flavoured slide pattern. For each concept the slides include:

1. **Motivation** — why the concept exists, what problem it solves.
2. **The basics** — a minimal code example.
3. **A worked example** — a substantive example using the concept in context.
4. **Common mistake** — what NOT to do, plus the bug the compiler/runtime surfaces.
5. **Recap** — a bullet list of takeaways.

The scaffolder's `slides.md.tmpl` pre-stocks three concept blocks following this pattern. Expand or contract per lesson.

Slide prose is self-readable. Use `Note:` blocks for live-lecture-only commentary.

### When students start writing tests

Through lessons 1-3, exercises ship pre-written failing tests that students don't author themselves — they just edit the `.go` files until the tests pass. From **lesson 4 onward**, students write some test code themselves.

### Imports

Phase 1 stays on the standard library only. The root `go.mod`'s `require` block stays empty through all of Phase 1. Third-party dependencies enter Phase 2 onward, and only when a lesson teaches them.

### Reference

- [Course design](docs/superpowers/specs/2026-05-05-go-course-design.md)
- [Phase 1 design (Plan C spec)](docs/superpowers/specs/2026-05-07-plan-c-phase-1-foundations-design.md)

````

- [ ] **Step 3: Confirm the file still parses as valid markdown**

A quick eyeball plus:

```bash
head -3 CONTRIBUTING.md
grep -c "^## " CONTRIBUTING.md
```

Expected:
- `head -3` still shows `# Contributing` followed by the Authoring section.
- The number of `## ` (level-2 heading) lines went up by 1 (the new "Phase 1 conventions" section).

- [ ] **Step 4: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "docs: document Phase 1 conventions in CONTRIBUTING"
```

---

## Task 5: End-to-end verification

Confirm the platform updates land cleanly: `make new-lesson` produces a 12-file skeleton, all three test commands work against the scaffolded output, and `git status` is clean afterward.

**Files:** none modified — verification only.

- [ ] **Step 1: Scaffold a sandbox lesson**

```bash
cd /Users/ristkari/code/private/go-training
make new-lesson NAME=99-demo
```

Expected: prints `created lesson 99-demo under lessons/`.

- [ ] **Step 2: Verify the 12-file tree**

```bash
find lessons/99-demo -type f | sort
```

Expected (12 lines):

```
lessons/99-demo/README.md
lessons/99-demo/exercises/main.go
lessons/99-demo/exercises/main_test.go
lessons/99-demo/exercises/warmup.go
lessons/99-demo/exercises/warmup_test.go
lessons/99-demo/slides/assets/.gitkeep
lessons/99-demo/slides/index.html
lessons/99-demo/slides/slides.md
lessons/99-demo/solutions/main.go
lessons/99-demo/solutions/main_test.go
lessons/99-demo/solutions/warmup.go
lessons/99-demo/solutions/warmup_test.go
```

- [ ] **Step 3: Verify the warm-up content was substituted**

```bash
grep -c "WarmupGreet" lessons/99-demo/exercises/warmup.go
grep -c "WarmupGreet" lessons/99-demo/solutions/warmup.go
```

Expected: both >0 (the warm-up function name is hardcoded in the templates and doesn't depend on substitution, but the file should exist and contain the function).

- [ ] **Step 4: Verify the README mentions warm-up and main**

```bash
grep -c "^## Exercise: warm-up$" lessons/99-demo/README.md
grep -c "^## Exercise: main$" lessons/99-demo/README.md
```

Expected: both `1`.

- [ ] **Step 5: Verify the slides skeleton has Concept blocks**

```bash
grep -c "^## Concept " lessons/99-demo/slides/slides.md
grep -c "^### Motivation$" lessons/99-demo/slides/slides.md
grep -c "^### Common mistake$" lessons/99-demo/slides/slides.md
```

Expected: `3`, `3`, `3` (three concept blocks, each with a Motivation and a Common mistake subsection).

- [ ] **Step 6: Run the demo lesson's exercise tests — must fail by design**

```bash
go test ./lessons/99-demo/exercises/... -v 2>&1 | tail -15
```

Expected: both `TestWarmupGreet` and `TestGreet` (the existing main exercise test from Plan A's template) fail. The output mentions panics about implementing `WarmupGreet` and `Greet`.

- [ ] **Step 7: Run the demo lesson's solution tests — must pass**

```bash
go test ./lessons/99-demo/solutions/... -v
```

Expected: both `TestWarmupGreet` and `TestGreet` PASS.

- [ ] **Step 8: Run the slides dev server briefly to confirm the deck renders**

```bash
make slides-dev LESSON=99-demo &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/99-demo/slides/slides.md
curl -sS http://localhost:8000/lessons/99-demo/slides/slides.md | grep -c "^## Concept "
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200` (the deck markdown is served), then `3` (the markdown contains 3 concept blocks).

- [ ] **Step 9: Run the full repo's `make test` — must pass**

```bash
make test
```

Expected: every package PASSes. `make test` excludes `*/exercises/*`, so the failing demo lesson exercise tests don't break the build.

- [ ] **Step 10: Run lint — must report 0 issues**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 11: Clean up the sandbox lesson**

```bash
rm -rf lessons/99-demo
test ! -d lessons/99-demo && echo OK
```

Expected: `OK`. Nothing to commit — the demo lesson was never staged.

- [ ] **Step 12: Final repository sanity check**

```bash
git status
make test
```

Expected: clean tree, all tests pass.

This task makes no commit — it is verification only.

---

## Done definition

After Task 5, all of these are true:

- `make new-lesson NAME=NN-name` produces a 12-file lesson skeleton (4 new warm-up files added on top of Plan A's 8).
- The scaffolded `slides/slides.md` follows the heavy-explanatory pattern with three concept blocks, each containing Motivation/Basics/Worked example/Common mistake/Recap subsections.
- The scaffolded `README.md` has separate "Exercise: warm-up" and "Exercise: main" sections plus a "How to run" with both commands.
- The scaffolded `exercises/warmup.go` defines `WarmupGreet` as a `panic("TODO")` stub; the matching `solutions/warmup.go` has it implemented; the matching `*_test.go` tests pass against the solution and fail against the exercise stub.
- `tools/new-lesson/main_test.go`'s `TestScaffoldCreatesExpectedTree` checks for the new 12-file layout.
- `CONTRIBUTING.md` has a "Phase 1 conventions" section documenting the warm-up + main pattern, the heavy-explanatory slide style, the lesson-4 testing handover, and the stdlib-only rule.
- `make test` passes; `golangci-lint run ./...` reports 0 issues.
- The git history is a clean sequence of small, conventional commits — one per task.

## What ships next (after Plan C merges)

- **Plan D** — Lesson 1 content (Hello, Go). Authored as a separate ~500-700-line plan with the literal slide-by-slide deck content, full README prose, exercise + solution + test code. Reviewed by the user before execution. Each subsequent lesson follows the same pattern (E, F, G, H, I, J, K).
