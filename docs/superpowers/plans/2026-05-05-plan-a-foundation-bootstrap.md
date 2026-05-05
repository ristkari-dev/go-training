# Plan A — Foundation Bootstrap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the `go-training` repo with a working Go module, vendored reveal.js, custom theme, lesson scaffolder, local slides dev server, Makefile, and contributor docs — so that `make new-lesson NAME=99-demo` produces a working lesson and `make slides-dev LESSON=99-demo` serves it locally.

**Architecture:** Single Go module at the root. Two small Go programs under `tools/`: `new-lesson` (generates a lesson folder from an embedded template) and `slides-dev` (static HTTP server that mounts the lesson's `slides/` plus the shared `reveal/` assets). Reveal.js 5.x is vendored under `shared/reveal/` and shared across all decks. A single Makefile is the canonical entry point for every workflow. No npm, no Node — Go + vendored static assets only.

**Tech Stack:** Go 1.23, reveal.js 5.1.0 (vendored), GNU Make, golangci-lint (developer machine), `goimports` (developer machine).

---

## File Structure

After this plan completes, the repo contains:

```
go-training/
├── README.md                              (Task 11)
├── CONTRIBUTING.md                        (Task 11)
├── go.mod                                 (Task 1)
├── go.sum                                 (Task 1, may be empty)
├── Makefile                               (Task 10)
├── .gitignore                             (Task 1, extends existing)
├── .editorconfig                          (Task 1)
├── .golangci.yml                          (Task 1)
│
├── shared/
│   └── reveal/                            (Task 2)
│       ├── LICENSE
│       ├── dist/                          (vendored reveal.js distributables)
│       ├── plugin/                        (vendored reveal.js plugins)
│       └── theme/
│           └── go-training.css            (Task 3, custom theme)
│
├── tools/
│   ├── new-lesson/
│   │   ├── main.go                        (Tasks 5-7)
│   │   ├── main_test.go                   (Tasks 5-7)
│   │   └── template/                      (Task 4)
│   │       ├── README.md.tmpl
│   │       ├── slides/
│   │       │   ├── index.html.tmpl
│   │       │   ├── slides.md.tmpl
│   │       │   └── assets/
│   │       │       └── .gitkeep
│   │       ├── exercises/
│   │       │   ├── main.go.tmpl
│   │       │   └── main_test.go.tmpl
│   │       └── solutions/
│   │           ├── main.go.tmpl
│   │           └── main_test.go.tmpl
│   └── slides-dev/
│       ├── main.go                        (Tasks 8-9)
│       └── main_test.go                   (Tasks 8-9)
│
└── docs/superpowers/specs/                (already exists)
└── docs/superpowers/plans/                (already exists)
```

### Decomposition rationale

- `new-lesson` is split into three single-responsibility units inside one package: `parseName` (input validation), `scaffold` (filesystem + template), and `main` (CLI wiring). Each is testable in isolation; the package stays small enough to read in one screen.
- `slides-dev` keeps its handler factory (`buildHandler`) separate from `main` so the handler is unit-testable via `httptest` without spinning up a real listener.
- The lesson template lives under `tools/new-lesson/template/` and is embedded at compile time via `//go:embed` — the binary needs no runtime filesystem dependency.

---

## Conventions used by this plan

- **Module path:** `github.com/<owner>/go-training`. **Before starting Task 1**, replace `<owner>` with the actual GitHub username/org. The placeholder appears only in Task 1.
- **Commit messages:** Conventional Commits (`feat:`, `chore:`, `docs:`, `test:`, `build:`).
- **Working directory:** `/Users/ristkari/code/private/go-training/` for every command. Do not `cd ..`.
- **Go version:** 1.23. If you want a different version, change it in Task 1 before continuing — it affects `go.mod` and (later, in Plan B) the Dockerfile.

---

## Task 1: Initialize Go module and base config files

**Files:**
- Create: `go.mod`
- Create: `.editorconfig`
- Create: `.golangci.yml`
- Modify: `.gitignore`

- [ ] **Step 1: Initialize the Go module**

```bash
cd /Users/ristkari/code/private/go-training
go mod init github.com/<owner>/go-training
```

Expected: creates `go.mod` containing `module github.com/<owner>/go-training` and `go 1.23` (or your installed version).

If your Go is newer than 1.23, edit `go.mod` so the `go` directive reads `go 1.23`. Pinning the floor keeps CI and local builds aligned.

- [ ] **Step 2: Extend `.gitignore`**

The repo already has a `.gitignore` from the design-doc commit. Append the following (use the Edit tool — do not overwrite):

```
# Go build artifacts
*.exe
*.test
*.out

# Editor
.idea/
.vscode/

# Slide build output (Plan B)
dist/
```

- [ ] **Step 3: Create `.editorconfig`**

Write `/Users/ristkari/code/private/go-training/.editorconfig`:

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2

[*.go]
indent_style = tab
indent_size = 4

[Makefile]
indent_style = tab
```

- [ ] **Step 4: Create `.golangci.yml`**

Write `/Users/ristkari/code/private/go-training/.golangci.yml`:

```yaml
run:
  timeout: 3m

linters:
  disable-all: true
  enable:
    - errcheck
    - gofmt
    - goimports
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused

issues:
  exclude-dirs:
    - shared/reveal
  exclude-rules:
    # Teaching examples often have intentionally simple main() functions.
    - path: lessons/.*/(exercises|solutions)/
      linters:
        - errcheck
```

- [ ] **Step 5: Verify the module builds**

```bash
go build ./...
```

Expected: completes silently with no output (no Go files exist yet, but `go build ./...` against an empty module succeeds).

- [ ] **Step 6: Commit**

```bash
git add go.mod .gitignore .editorconfig .golangci.yml
git commit -m "chore: initialize Go module and base config"
```

---

## Task 2: Vendor reveal.js 5.1.0

**Files:**
- Create: `shared/reveal/` (entire tree downloaded from upstream)

- [ ] **Step 1: Download the upstream tarball into a scratch directory**

```bash
mkdir -p /tmp/reveal-vendor && cd /tmp/reveal-vendor
curl -fsSL -o reveal.tar.gz https://github.com/hakimel/reveal.js/archive/refs/tags/5.1.0.tar.gz
tar -xzf reveal.tar.gz
ls reveal.js-5.1.0
```

Expected: a directory `reveal.js-5.1.0/` listing `dist/`, `plugin/`, `LICENSE`, `package.json`, etc.

- [ ] **Step 2: Copy only the parts we need into the repo**

```bash
mkdir -p /Users/ristkari/code/private/go-training/shared/reveal
cp -R /tmp/reveal-vendor/reveal.js-5.1.0/dist   /Users/ristkari/code/private/go-training/shared/reveal/
cp -R /tmp/reveal-vendor/reveal.js-5.1.0/plugin /Users/ristkari/code/private/go-training/shared/reveal/
cp    /tmp/reveal-vendor/reveal.js-5.1.0/LICENSE /Users/ristkari/code/private/go-training/shared/reveal/
```

- [ ] **Step 3: Record the version**

Write `/Users/ristkari/code/private/go-training/shared/reveal/VERSION`:

```
reveal.js 5.1.0
Source: https://github.com/hakimel/reveal.js/releases/tag/5.1.0
Vendored: 2026-05-05
Upgrade: re-download the tarball, replace dist/ and plugin/, update this file.
```

- [ ] **Step 4: Sanity-check the vendored content**

```bash
test -f /Users/ristkari/code/private/go-training/shared/reveal/dist/reveal.js && echo OK
test -f /Users/ristkari/code/private/go-training/shared/reveal/dist/reveal.css && echo OK
test -d /Users/ristkari/code/private/go-training/shared/reveal/plugin/markdown && echo OK
test -d /Users/ristkari/code/private/go-training/shared/reveal/plugin/highlight && echo OK
test -d /Users/ristkari/code/private/go-training/shared/reveal/plugin/notes && echo OK
test -d /Users/ristkari/code/private/go-training/shared/reveal/plugin/search && echo OK
```

Expected: six lines of `OK`.

- [ ] **Step 5: Clean up the scratch directory**

```bash
rm -rf /tmp/reveal-vendor
```

- [ ] **Step 6: Commit**

```bash
git add shared/reveal
git commit -m "build: vendor reveal.js 5.1.0"
```

---

## Task 3: Create the custom reveal.js theme

**Files:**
- Create: `shared/reveal/theme/go-training.css`

- [ ] **Step 1: Write the theme**

Write `/Users/ristkari/code/private/go-training/shared/reveal/theme/go-training.css`:

```css
/*
 * go-training: a reveal.js theme tuned for code-heavy slides.
 * Builds on reveal.js's CSS variables; override only what we need.
 */

:root {
  --r-background-color: #1f2430;
  --r-main-color: #f4f4f4;
  --r-main-font: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  --r-main-font-size: 38px;
  --r-heading-color: #5dc9e2;
  --r-heading-font: var(--r-main-font);
  --r-heading-text-shadow: none;
  --r-heading-text-transform: none;
  --r-heading1-size: 2.2em;
  --r-heading2-size: 1.6em;
  --r-heading3-size: 1.2em;
  --r-link-color: #5dc9e2;
  --r-link-color-hover: #ffffff;
  --r-selection-color: #1f2430;
  --r-selection-background-color: #5dc9e2;
  --r-code-font: "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
}

.reveal {
  background-color: var(--r-background-color);
  color: var(--r-main-color);
}

.reveal h1, .reveal h2, .reveal h3, .reveal h4 {
  color: var(--r-heading-color);
  letter-spacing: -0.01em;
}

.reveal pre {
  width: 100%;
  font-size: 0.7em;
  box-shadow: none;
  border-radius: 6px;
}

.reveal pre code {
  padding: 1em 1.2em;
  max-height: 600px;
  line-height: 1.4;
  font-family: var(--r-code-font);
}

.reveal code {
  font-family: var(--r-code-font);
  background: rgba(255, 255, 255, 0.08);
  padding: 0.05em 0.3em;
  border-radius: 3px;
}

.reveal section img {
  border: none;
  background: transparent;
  box-shadow: none;
}

.reveal blockquote {
  background: rgba(255, 255, 255, 0.05);
  border-left: 4px solid var(--r-heading-color);
  padding: 0.5em 1em;
  font-style: normal;
}

.reveal .slide-number {
  font-family: var(--r-code-font);
  background: transparent;
  color: rgba(255, 255, 255, 0.4);
}
```

- [ ] **Step 2: Commit**

```bash
git add shared/reveal/theme/go-training.css
git commit -m "feat(slides): add go-training reveal.js theme"
```

---

## Task 4: Create the lesson template files

These files become the bytes that `tools/new-lesson` embeds and renders. The `.tmpl` suffix is stripped on output and the contents go through Go's `text/template` engine. The template engine receives a struct with these fields:

- `.Number` — two-digit lesson number (e.g. `"01"`)
- `.Slug`   — kebab-case slug (e.g. `"slices-and-maps"`)
- `.Name`   — full lesson name (e.g. `"01-hello"`)
- `.Title`  — slug with first letter of each word capitalised and dashes replaced with spaces (e.g. `"Slices And Maps"`)

> The `TODO` markers in these template files are intentional placeholders for the *lesson author* to fill in when authoring a real lesson. They are not implementation placeholders for this plan.

**Files:**
- Create: `tools/new-lesson/template/README.md.tmpl`
- Create: `tools/new-lesson/template/slides/index.html.tmpl`
- Create: `tools/new-lesson/template/slides/slides.md.tmpl`
- Create: `tools/new-lesson/template/slides/assets/.gitkeep`
- Create: `tools/new-lesson/template/exercises/main.go.tmpl`
- Create: `tools/new-lesson/template/exercises/main_test.go.tmpl`
- Create: `tools/new-lesson/template/solutions/main.go.tmpl`
- Create: `tools/new-lesson/template/solutions/main_test.go.tmpl`

- [ ] **Step 1: Write `tools/new-lesson/template/README.md.tmpl`**

```markdown
# Lesson {{.Number}}: {{.Title}}

## Learning goals

- TODO
- TODO
- TODO

## Prerequisites

- Lessons through {{.Number}} (this one is built on the prior lesson's material).

## Concepts

TODO — write the prose narrative that mirrors the slides for self-study readers.

## Exercise

Open `exercises/`. There is starter code that compiles plus failing tests in `*_test.go`. Make the tests pass.

## How to run

```bash
cd lessons/{{.Name}}/exercises
go test ./...
go run .
```

## Going further

- TODO — links, deeper exercises, std-lib reading for advanced students.
```

- [ ] **Step 2: Write `tools/new-lesson/template/slides/index.html.tmpl`**

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
  <title>Lesson {{.Number}} — {{.Title}}</title>
  <link rel="stylesheet" href="../../../shared/reveal/dist/reset.css">
  <link rel="stylesheet" href="../../../shared/reveal/dist/reveal.css">
  <link rel="stylesheet" href="../../../shared/reveal/theme/go-training.css">
  <link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/monokai.css">
</head>
<body>
  <div class="reveal">
    <div class="slides">
      <section data-markdown="slides.md"
               data-separator="^---$"
               data-separator-vertical="^--$"
               data-separator-notes="^Note:"></section>
    </div>
  </div>
  <script src="../../../shared/reveal/dist/reveal.js"></script>
  <script src="../../../shared/reveal/plugin/markdown/markdown.js"></script>
  <script src="../../../shared/reveal/plugin/highlight/highlight.js"></script>
  <script src="../../../shared/reveal/plugin/notes/notes.js"></script>
  <script src="../../../shared/reveal/plugin/search/search.js"></script>
  <script>
    Reveal.initialize({
      hash: true,
      slideNumber: 'c/t',
      plugins: [RevealMarkdown, RevealHighlight, RevealNotes, RevealSearch],
    });
  </script>
</body>
</html>
```

- [ ] **Step 3: Write `tools/new-lesson/template/slides/slides.md.tmpl`**

```markdown
## Lesson {{.Number}}

# {{.Title}}

Learning goal: TODO

---

## What we'll cover

- TODO
- TODO
- TODO

---

## A first example

```go
package main

import "fmt"

func main() {
	fmt.Println("TODO")
}
```

Note:
Speaker notes for the live lecture go here. They render only in the speaker view (press `S`).

---

## Recap

- TODO
- TODO

---

## Up next

See the next lesson under `lessons/`.
```

- [ ] **Step 4: Write `tools/new-lesson/template/slides/assets/.gitkeep`**

Empty file — keeps the directory tracked in git.

- [ ] **Step 5: Write `tools/new-lesson/template/exercises/main.go.tmpl`**

```go
// Package exercises is the starter code for lesson {{.Number}}: {{.Title}}.
//
// Make the failing tests in main_test.go pass.
package exercises

// Greet returns a greeting for name, e.g. Greet("World") == "Hello, World!".
func Greet(name string) string {
	panic("TODO: implement Greet")
}
```

- [ ] **Step 6: Write `tools/new-lesson/template/exercises/main_test.go.tmpl`**

```go
package exercises

import "testing"

func TestGreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"World", "Hello, World!"},
		{"Go", "Hello, Go!"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := Greet(tc.in)
			if got != tc.want {
				t.Errorf("Greet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 7: Write `tools/new-lesson/template/solutions/main.go.tmpl`**

```go
// Package solutions is the reference implementation for lesson {{.Number}}: {{.Title}}.
package solutions

// Greet returns a greeting for name.
func Greet(name string) string {
	return "Hello, " + name + "!"
}
```

- [ ] **Step 8: Write `tools/new-lesson/template/solutions/main_test.go.tmpl`**

```go
package solutions

import "testing"

func TestGreet(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"World", "Hello, World!"},
		{"Go", "Hello, Go!"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := Greet(tc.in)
			if got != tc.want {
				t.Errorf("Greet(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 9: Verify the template tree**

```bash
find /Users/ristkari/code/private/go-training/tools/new-lesson/template -type f | sort
```

Expected (8 lines):

```
/Users/ristkari/code/private/go-training/tools/new-lesson/template/README.md.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/exercises/main.go.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/exercises/main_test.go.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/slides/assets/.gitkeep
/Users/ristkari/code/private/go-training/tools/new-lesson/template/slides/index.html.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/slides/slides.md.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/solutions/main.go.tmpl
/Users/ristkari/code/private/go-training/tools/new-lesson/template/solutions/main_test.go.tmpl
```

- [ ] **Step 10: Commit**

```bash
git add tools/new-lesson/template
git commit -m "feat(new-lesson): add lesson template files"
```

---

## Task 5: Implement `parseName` for the scaffolder (TDD)

`parseName` validates and parses a lesson name like `"05-slices-and-maps"` into a `lessonInfo` struct that the scaffolder uses for template substitution.

**Files:**
- Create: `tools/new-lesson/main.go`
- Create: `tools/new-lesson/main_test.go`

- [ ] **Step 1: Write the failing test**

Create `/Users/ristkari/code/private/go-training/tools/new-lesson/main_test.go`:

```go
package main

import (
	"testing"
)

func TestParseName(t *testing.T) {
	t.Run("valid simple", func(t *testing.T) {
		got, err := parseName("01-hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lessonInfo{Number: "01", Slug: "hello", Name: "01-hello", Title: "Hello"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("valid multi-word", func(t *testing.T) {
		got, err := parseName("05-slices-and-maps")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := lessonInfo{Number: "05", Slug: "slices-and-maps", Name: "05-slices-and-maps", Title: "Slices And Maps"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("rejects invalid forms", func(t *testing.T) {
		bad := []string{
			"",
			"hello",
			"1-hello",
			"001-hello",
			"01_hello",
			"01-Hello",
			"01-hello-",
			"-hello",
			"01-",
		}
		for _, name := range bad {
			if _, err := parseName(name); err == nil {
				t.Errorf("expected error for %q", name)
			}
		}
	})
}
```

- [ ] **Step 2: Run the test, confirm it fails**

```bash
cd /Users/ristkari/code/private/go-training
go test ./tools/new-lesson/ -run TestParseName -v
```

Expected: build failure — `parseName` and `lessonInfo` are not defined.

- [ ] **Step 3: Write the minimal implementation**

Create `/Users/ristkari/code/private/go-training/tools/new-lesson/main.go`:

```go
// Command new-lesson scaffolds a new go-training lesson folder from
// the embedded template under tools/new-lesson/template.
package main

import (
	"fmt"
	"regexp"
	"strings"
)

type lessonInfo struct {
	Number string // "01"
	Slug   string // "hello"
	Name   string // "01-hello"
	Title  string // "Hello"
}

var nameRe = regexp.MustCompile(`^(\d{2})-([a-z][a-z0-9]*(?:-[a-z0-9]+)*)$`)

func parseName(name string) (lessonInfo, error) {
	m := nameRe.FindStringSubmatch(name)
	if m == nil {
		return lessonInfo{}, fmt.Errorf("invalid lesson name %q: must match NN-kebab-case, e.g. 01-hello", name)
	}
	return lessonInfo{
		Number: m[1],
		Slug:   m[2],
		Name:   name,
		Title:  toTitle(m[2]),
	}, nil
}

func toTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func main() {
	// Wired in Task 7.
}
```

- [ ] **Step 4: Run the test, confirm it passes**

```bash
go test ./tools/new-lesson/ -run TestParseName -v
```

Expected: `--- PASS: TestParseName` with three sub-tests passing.

- [ ] **Step 5: Commit**

```bash
git add tools/new-lesson/main.go tools/new-lesson/main_test.go
git commit -m "feat(new-lesson): add parseName with validation"
```

---

## Task 6: Implement scaffold (file tree generation + template substitution) (TDD)

`scaffold` walks the embedded template filesystem and writes the rendered output to a destination directory. `.tmpl` files are processed through `text/template`; non-`.tmpl` files are copied verbatim. It refuses to overwrite an existing destination.

**Files:**
- Modify: `tools/new-lesson/main.go`
- Modify: `tools/new-lesson/main_test.go`

- [ ] **Step 1: Add the failing tests**

First, replace the existing `import "testing"` line in `tools/new-lesson/main_test.go` with a parenthesised group containing all four imports:

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

Then append the three new test functions to the end of the file:

```go
func TestScaffoldCreatesExpectedTree(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("01-hello", dest); err != nil {
		t.Fatalf("scaffoldLesson: %v", err)
	}
	want := []string{
		"01-hello/README.md",
		"01-hello/slides/index.html",
		"01-hello/slides/slides.md",
		"01-hello/slides/assets/.gitkeep",
		"01-hello/exercises/main.go",
		"01-hello/exercises/main_test.go",
		"01-hello/solutions/main.go",
		"01-hello/solutions/main_test.go",
	}
	for _, rel := range want {
		p := filepath.Join(dest, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestScaffoldSubstitutesLessonInfo(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("05-slices-and-maps", dest); err != nil {
		t.Fatalf("scaffoldLesson: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(dest, "05-slices-and-maps", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Lesson 05", "Slices And Maps", "lessons/05-slices-and-maps/exercises"} {
		if !strings.Contains(string(readme), want) {
			t.Errorf("README missing %q\n---\n%s", want, readme)
		}
	}
}

func TestScaffoldRefusesExistingDest(t *testing.T) {
	dest := t.TempDir()
	if err := scaffoldLesson("01-hello", dest); err != nil {
		t.Fatalf("first scaffoldLesson: %v", err)
	}
	err := scaffoldLesson("01-hello", dest)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected already-exists error, got %v", err)
	}
}
```

- [ ] **Step 2: Run the tests, confirm they fail**

```bash
go test ./tools/new-lesson/ -v
```

Expected: build failure — `scaffoldLesson` is not defined.

- [ ] **Step 3: Implement scaffold**

Replace `tools/new-lesson/main.go` with the complete implementation below. The previous `parseName` and `toTitle` are kept; `scaffoldLesson` and template embedding are added; `main` stays empty for now (wired in Task 7).

```go
// Command new-lesson scaffolds a new go-training lesson folder from
// the embedded template under tools/new-lesson/template.
package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed all:template
var templateFS embed.FS

type lessonInfo struct {
	Number string
	Slug   string
	Name   string
	Title  string
}

var nameRe = regexp.MustCompile(`^(\d{2})-([a-z][a-z0-9]*(?:-[a-z0-9]+)*)$`)

func parseName(name string) (lessonInfo, error) {
	m := nameRe.FindStringSubmatch(name)
	if m == nil {
		return lessonInfo{}, fmt.Errorf("invalid lesson name %q: must match NN-kebab-case, e.g. 01-hello", name)
	}
	return lessonInfo{
		Number: m[1],
		Slug:   m[2],
		Name:   name,
		Title:  toTitle(m[2]),
	}, nil
}

func toTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// scaffoldLesson creates a new lesson under outBase using the embedded template.
// It refuses to overwrite an existing destination.
func scaffoldLesson(name, outBase string) error {
	info, err := parseName(name)
	if err != nil {
		return err
	}
	dest := filepath.Join(outBase, info.Name)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination already exists: %s", dest)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return walkTemplate(templateFS, "template", dest, info)
}

func walkTemplate(srcFS fs.FS, root, dest string, info lessonInfo) error {
	return fs.WalkDir(srcFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dest, 0o755)
		}
		target := filepath.Join(dest, strings.TrimSuffix(rel, ".tmpl"))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(srcFS, path)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return os.WriteFile(target, data, 0o644)
		}
		tmpl, err := template.New(rel).Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", rel, err)
		}
		f, err := os.Create(target)
		if err != nil {
			return err
		}
		defer f.Close()
		return tmpl.Execute(f, info)
	})
}

func main() {
	// Wired in Task 7.
}
```

- [ ] **Step 4: Run the tests, confirm they pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: every `TestParseName` sub-test plus `TestScaffoldCreatesExpectedTree`, `TestScaffoldSubstitutesLessonInfo`, and `TestScaffoldRefusesExistingDest` all PASS.

- [ ] **Step 5: Commit**

```bash
git add tools/new-lesson/main.go tools/new-lesson/main_test.go
git commit -m "feat(new-lesson): scaffold lesson folder from embedded template"
```

---

## Task 7: Wire the scaffolder CLI

Wires `main()` to parse flags, call `scaffoldLesson`, and exit with a non-zero code on error. A small CLI integration test exercises the binary's exit behaviour.

**Files:**
- Modify: `tools/new-lesson/main.go`
- Modify: `tools/new-lesson/main_test.go`

- [ ] **Step 1: Add the CLI integration test**

Append to `tools/new-lesson/main_test.go`:

```go
func TestCLIRejectsMissingName(t *testing.T) {
	if code := runWithArgs([]string{"new-lesson"}); code == 0 {
		t.Fatal("expected non-zero exit when -name is missing")
	}
}

func TestCLIScaffoldsHappyPath(t *testing.T) {
	dest := t.TempDir()
	code := runWithArgs([]string{"new-lesson", "-name", "07-packages", "-out", dest})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(dest, "07-packages", "README.md")); err != nil {
		t.Fatalf("README not created: %v", err)
	}
}
```

- [ ] **Step 2: Run the tests, confirm they fail**

```bash
go test ./tools/new-lesson/ -v
```

Expected: build failure — `runWithArgs` is not defined.

- [ ] **Step 3: Wire main()**

Replace the `func main()` stub at the bottom of `tools/new-lesson/main.go` with:

```go
func main() {
	os.Exit(runWithArgs(os.Args))
}

// runWithArgs parses args[1:] as flags and runs the scaffolder.
// It returns the desired process exit code so it can be unit-tested.
func runWithArgs(args []string) int {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	name := fs.String("name", "", "lesson name in NN-kebab-case (e.g. 01-hello)")
	out := fs.String("out", "lessons", "output base directory")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required (e.g. -name 01-hello)")
		return 2
	}
	if err := scaffoldLesson(*name, *out); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Printf("created lesson %s under %s/\n", *name, *out)
	return 0
}
```

Add `"flag"` to the existing `import (…)` block.

- [ ] **Step 4: Run the tests, confirm they pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: all five tests PASS (`TestParseName`, `TestScaffold*` × 3, `TestCLI*` × 2).

- [ ] **Step 5: Smoke-test the binary directly**

```bash
TMP=$(mktemp -d)
go run ./tools/new-lesson -name 99-demo -out "$TMP"
ls "$TMP/99-demo"
go test "./$TMP/99-demo/solutions/..." 2>&1 || true
rm -rf "$TMP"
```

Expected: `ls` shows `README.md exercises/ slides/ solutions/`. `go test` step is informational; it may not run cleanly outside the module — that's fine, it's exercised properly inside the module in Task 12.

- [ ] **Step 6: Commit**

```bash
git add tools/new-lesson/main.go tools/new-lesson/main_test.go
git commit -m "feat(new-lesson): wire CLI with -name and -out flags"
```

---

## Task 8: Implement the slides-dev HTTP handler (TDD)

`buildHandler(repo, lesson)` returns an `http.Handler` that serves:

- `/lessons/...` and `/shared/...` directly from the repo (so that the lesson's `index.html`, which references `../../../shared/reveal/...`, resolves).
- `/` redirects to the lesson's deck URL.
- Unknown paths → 404.

It also validates that the requested lesson and the shared reveal directory exist on disk.

**Files:**
- Create: `tools/slides-dev/main.go`
- Create: `tools/slides-dev/main_test.go`

- [ ] **Step 1: Write the failing tests**

Create `/Users/ristkari/code/private/go-training/tools/slides-dev/main_test.go`:

```go
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/slides.md"), "# Hello\n")
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/index.html"), "<html>hi</html>")
	mustWrite(t, filepath.Join(dir, "shared/reveal/dist/reveal.js"), "console.log('reveal');")
	return dir
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHandlerServesLessonSlides(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/lessons/01-hello/slides/slides.md", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello") {
		t.Errorf("body: %q does not contain 'Hello'", rec.Body.String())
	}
}

func TestHandlerServesRevealAssets(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/shared/reveal/dist/reveal.js", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
}

func TestHandlerRedirectsRoot(t *testing.T) {
	repo := setupRepo(t)
	h, err := buildHandler(repo, "01-hello")
	if err != nil {
		t.Fatalf("buildHandler: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status: got %d, want 302", rec.Code)
	}
	want := "/lessons/01-hello/slides/"
	if got := rec.Header().Get("Location"); got != want {
		t.Errorf("Location: got %q, want %q", got, want)
	}
}

func TestHandlerErrorsOnMissingLesson(t *testing.T) {
	repo := setupRepo(t)
	if _, err := buildHandler(repo, "99-nope"); err == nil {
		t.Error("expected error for missing lesson")
	}
}

func TestHandlerErrorsOnMissingReveal(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "lessons/01-hello/slides/slides.md"), "# Hi")
	if _, err := buildHandler(dir, "01-hello"); err == nil {
		t.Error("expected error when shared/reveal is absent")
	}
}
```

- [ ] **Step 2: Run the tests, confirm they fail**

```bash
go test ./tools/slides-dev/ -v
```

Expected: build failure — `buildHandler` is not defined.

- [ ] **Step 3: Implement the handler**

Create `/Users/ristkari/code/private/go-training/tools/slides-dev/main.go`:

```go
// Command slides-dev serves a single lesson's reveal.js deck over HTTP for local
// development. It mounts the repo's lessons/ and shared/ trees so relative paths
// inside the lesson's index.html resolve the same way they do on disk.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// buildHandler returns an http.Handler that serves lessons and shared assets
// from the repo, and redirects "/" to the requested lesson's deck.
func buildHandler(repo, lesson string) (http.Handler, error) {
	slidesDir := filepath.Join(repo, "lessons", lesson, "slides")
	if _, err := os.Stat(slidesDir); err != nil {
		return nil, fmt.Errorf("lesson slides not found at %s: %w", slidesDir, err)
	}
	revealDir := filepath.Join(repo, "shared", "reveal")
	if _, err := os.Stat(revealDir); err != nil {
		return nil, fmt.Errorf("shared reveal not found at %s: %w", revealDir, err)
	}

	deckURL := "/lessons/" + lesson + "/slides/"
	repoFS := http.FileServer(http.Dir(repo))

	mux := http.NewServeMux()
	mux.Handle("/lessons/", repoFS)
	mux.Handle("/shared/", repoFS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, deckURL, http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	return mux, nil
}

func main() {
	// Wired in Task 9.
}
```

- [ ] **Step 4: Run the tests, confirm they pass**

```bash
go test ./tools/slides-dev/ -v
```

Expected: all five tests PASS.

- [ ] **Step 5: Commit**

```bash
git add tools/slides-dev/main.go tools/slides-dev/main_test.go
git commit -m "feat(slides-dev): add HTTP handler for serving lesson decks"
```

---

## Task 9: Wire the slides-dev CLI

**Files:**
- Modify: `tools/slides-dev/main.go`

- [ ] **Step 1: Replace the `main()` stub**

Replace `func main()` in `tools/slides-dev/main.go` with:

```go
func main() {
	os.Exit(runWithArgs(os.Args))
}

func runWithArgs(args []string) int {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	lesson := fs.String("lesson", "", "lesson name (e.g. 01-hello)")
	addr := fs.String("addr", ":8000", "listen address")
	repo := fs.String("repo", ".", "repo root containing lessons/ and shared/")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if *lesson == "" {
		fmt.Fprintln(os.Stderr, "error: -lesson is required (e.g. -lesson 01-hello)")
		return 2
	}
	h, err := buildHandler(*repo, *lesson)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	displayAddr := *addr
	if len(displayAddr) > 0 && displayAddr[0] == ':' {
		displayAddr = "localhost" + displayAddr
	}
	fmt.Printf("serving lesson %s on http://%s/  (Ctrl-C to stop)\n", *lesson, displayAddr)
	if err := http.ListenAndServe(*addr, h); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
```

Add `"flag"` to the existing `import (…)` block.

- [ ] **Step 2: Confirm tests still pass**

```bash
go test ./tools/slides-dev/ -v
```

Expected: all tests still PASS (no behavioural change).

- [ ] **Step 3: Commit**

```bash
git add tools/slides-dev/main.go
git commit -m "feat(slides-dev): wire CLI with -lesson, -addr, -repo flags"
```

---

## Task 10: Write the Makefile

A self-documenting Makefile that is the single canonical entry point for every workflow. Targets that belong to Plan B (`slides-build`, `slides-docker`) are intentionally absent.

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Write the Makefile**

Create `/Users/ristkari/code/private/go-training/Makefile`:

```make
SHELL := /bin/bash
.DEFAULT_GOAL := help

REPO_ROOT := $(shell pwd)

.PHONY: help
help: ## List available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: test
test: ## Run tests, excluding the intentionally-failing exercise tests
	@pkgs=$$(go list ./... | grep -v '/exercises$$'); \
	if [ -z "$$pkgs" ]; then echo "no testable packages"; exit 0; fi; \
	go test $$pkgs

.PHONY: test-exercises
test-exercises: ## Run exercise tests (these fail by design until students complete them)
	-@pkgs=$$(go list ./... | grep '/exercises$$'); \
	if [ -z "$$pkgs" ]; then echo "no exercise packages"; exit 0; fi; \
	go test $$pkgs

.PHONY: test-lesson
test-lesson: ## Run tests for one lesson, both exercises and solutions (LESSON=NN-name)
	@test -n "$(LESSON)" || (echo "usage: make test-lesson LESSON=NN-name" && exit 1)
	-go test ./lessons/$(LESSON)/exercises/...
	go test ./lessons/$(LESSON)/solutions/...

.PHONY: lint
lint: ## Run golangci-lint (must be installed locally)
	golangci-lint run

.PHONY: fmt
fmt: ## Format Go code with gofmt and goimports
	gofmt -w .
	goimports -w .

.PHONY: new-lesson
new-lesson: ## Scaffold a new lesson (NAME=NN-name)
	@test -n "$(NAME)" || (echo "usage: make new-lesson NAME=NN-name" && exit 1)
	go run ./tools/new-lesson -name $(NAME)

.PHONY: slides-dev
slides-dev: ## Serve one lesson's deck locally on http://localhost:8000 (LESSON=NN-name)
	@test -n "$(LESSON)" || (echo "usage: make slides-dev LESSON=NN-name" && exit 1)
	go run ./tools/slides-dev -lesson $(LESSON) -repo $(REPO_ROOT)
```

- [ ] **Step 2: Verify `make help` works**

```bash
make help
```

Expected: prints the seven targets above with their descriptions, each line beginning with two spaces and a coloured target name.

- [ ] **Step 3: Verify `make test` works**

```bash
make test
```

Expected: runs `go test` on the two tool packages (`./tools/new-lesson` and `./tools/slides-dev`), all PASS.

- [ ] **Step 4: Commit**

```bash
git add Makefile
git commit -m "build: add canonical Makefile"
```

---

## Task 11: Write top-level README and CONTRIBUTING

**Files:**
- Create: `README.md`
- Create: `CONTRIBUTING.md`

- [ ] **Step 1: Write `README.md`**

Create `/Users/ristkari/code/private/go-training/README.md`:

````markdown
# Go Training

A Go programming course delivered as code + per-lesson reveal.js slide decks.
The arc starts at programming-101 and finishes with concurrency, systems
programming, production services, tooling, and distributed patterns.

## Prerequisites

- Go 1.23 or newer (`go version`)
- Make
- `golangci-lint` and `goimports` for the dev workflow:
  ```bash
  go install golang.org/x/tools/cmd/goimports@latest
  brew install golangci-lint   # or see https://golangci-lint.run/welcome/install/
  ```

## Quick start

Clone the repo, then from the repo root:

```bash
make help                       # list every available command
make new-lesson NAME=99-demo    # scaffold a sandbox lesson
make slides-dev LESSON=99-demo  # serve its deck on http://localhost:8000
make test                       # run all tests except the intentionally-failing exercises
```

## Repository layout

```
lessons/NN-name/
├── README.md      self-study notes for the lesson
├── slides/        reveal.js deck (index.html + slides.md)
├── exercises/     starter code + failing tests (the spec)
└── solutions/     reference implementation

shared/reveal/     vendored reveal.js + custom theme (do not edit by hand)
tools/             developer tooling (new-lesson, slides-dev)
docs/              design docs and implementation plans
```

## Design

See [`docs/superpowers/specs/2026-05-05-go-course-design.md`](docs/superpowers/specs/2026-05-05-go-course-design.md)
for the course design.
````

- [ ] **Step 2: Write `CONTRIBUTING.md`**

Create `/Users/ristkari/code/private/go-training/CONTRIBUTING.md`:

````markdown
# Contributing

## Authoring a new lesson

```bash
make new-lesson NAME=15-channels
```

This creates `lessons/15-channels/` with the four-part structure: `README.md`,
`slides/`, `exercises/`, `solutions/`. The scaffolded files have `TODO`
markers — fill them in.

### The four-part structure

Every lesson has exactly these four parts:

1. **`README.md`** — self-study notes that mirror the deck narrative.
   Sections: Learning goals, Prerequisites, Concepts, Exercise, How to run,
   Going further.
2. **`slides/`** — the live-lecture deck. `index.html` is reveal.js bootstrap;
   `slides.md` is the markdown content. Use `Note:` blocks for speaker notes.
3. **`exercises/`** — starter code that **compiles** but is incomplete (use
   `panic("TODO: …")` for unimplemented bodies). The accompanying `*_test.go`
   files contain **failing tests** that act as the spec.
4. **`solutions/`** — the same package shape as `exercises/`, fully
   implemented. The tests in `solutions/` must be identical to those in
   `exercises/` so `go test ./lessons/NN-name/solutions/...` passes.

### Slide style

- First slide: lesson number, title, one-line learning goal.
- Last slide: pointer to the next lesson.
- Code-heavy slides: limit to ~15 visible lines. Split larger examples and
  use the highlight plugin's `[highlight]` syntax to focus attention.
- Diagrams: SVG. Never images of code.

### Tests as the spec

Exercise tests fail by design until the student completes the lesson.
`make test` excludes `*/exercises/*` so CI stays green; students opt in with
`make test-lesson LESSON=NN-name` (which runs both exercise and solution
tests, ignoring exercise failures).

## Local workflow

```bash
make fmt                        # format
make lint                       # lint
make test                       # all tests except exercises
make test-lesson LESSON=01-hello  # one lesson, both sides
make slides-dev LESSON=01-hello   # serve deck on http://localhost:8000
```

## Reveal.js

Reveal.js is vendored under `shared/reveal/`. To upgrade:

1. Download a new release tarball from
   <https://github.com/hakimel/reveal.js/releases>.
2. Replace `shared/reveal/dist/` and `shared/reveal/plugin/`.
3. Update `shared/reveal/VERSION`.
4. Smoke-test one lesson with `make slides-dev`.

The custom theme lives at `shared/reveal/theme/go-training.css` — keep it
across upgrades.

## Commit messages

Conventional Commits: `feat:`, `fix:`, `chore:`, `docs:`, `test:`, `build:`.
Scope is optional, e.g. `feat(slides): …`.
````

- [ ] **Step 3: Commit**

```bash
git add README.md CONTRIBUTING.md
git commit -m "docs: add top-level README and CONTRIBUTING"
```

---

## Task 12: End-to-end smoke test

Verifies the whole bootstrap works as a system: scaffold a lesson, run its solution tests, run its exercise tests (which must fail), serve its deck, and clean up.

**Files:** none modified — this task is verification only.

- [ ] **Step 1: Scaffold a sandbox lesson**

```bash
cd /Users/ristkari/code/private/go-training
make new-lesson NAME=99-demo
```

Expected: prints `created lesson 99-demo under lessons/`. Verify the tree:

```bash
find lessons/99-demo -type f | sort
```

Expected (8 files):

```
lessons/99-demo/README.md
lessons/99-demo/exercises/main.go
lessons/99-demo/exercises/main_test.go
lessons/99-demo/slides/assets/.gitkeep
lessons/99-demo/slides/index.html
lessons/99-demo/slides/slides.md
lessons/99-demo/solutions/main.go
lessons/99-demo/solutions/main_test.go
```

- [ ] **Step 2: Verify `make test` passes (excludes exercises)**

```bash
make test
```

Expected: passes. Output mentions `./lessons/99-demo/solutions`, `./tools/new-lesson`, `./tools/slides-dev` — and **does not** mention `./lessons/99-demo/exercises`.

- [ ] **Step 3: Verify the exercise tests fail by design**

```bash
make test-exercises
```

Expected: reports a failing test in `./lessons/99-demo/exercises` (the `panic("TODO: implement Greet")` panics during `TestGreet`). Make exits 0 because the recipe uses `-` to ignore the failure — the failure itself is the success criterion of this step.

- [ ] **Step 4: Verify `make test-lesson` works for both sides**

```bash
make test-lesson LESSON=99-demo
```

Expected: exercise tests fail (ignored), solution tests pass.

- [ ] **Step 5: Verify `make slides-dev` serves the deck**

In one terminal (or background it):

```bash
make slides-dev LESSON=99-demo &
SDPID=$!
sleep 2
```

In another:

```bash
curl -sS -o /dev/null -w "%{http_code} %{redirect_url}\n" http://localhost:8000/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/99-demo/slides/index.html
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/99-demo/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/dist/reveal.js
```

Expected:

- Line 1: `302 /lessons/99-demo/slides/`
- Lines 2-4: `200`

Stop the server:

```bash
kill $SDPID
wait $SDPID 2>/dev/null
```

- [ ] **Step 6: Visual check (optional, no automation)**

In a browser, open `http://localhost:8000/` while the dev server is running. You should see the demo deck rendered with the `go-training` dark theme. Press `Esc` for the slide overview, `S` for speaker notes.

- [ ] **Step 7: Clean up the sandbox lesson**

```bash
rm -rf lessons/99-demo
test ! -d lessons/99-demo && echo OK
```

Expected: `OK`. Nothing to commit — the sandbox lesson was never staged.

- [ ] **Step 8: Final repository sanity check**

```bash
git status
make test
```

Expected: `git status` shows a clean working tree. `make test` passes (only the tools tests run now that the demo lesson is gone).

---

## Done definition

After Task 12, all of these are true:

- `make help` lists seven targets and their descriptions.
- `make test` passes from a clean checkout.
- `make new-lesson NAME=…` creates a working lesson skeleton from the embedded template.
- `make slides-dev LESSON=…` serves a lesson deck over HTTP, with the `shared/reveal/` assets available at `/shared/reveal/`.
- `shared/reveal/` contains the vendored reveal.js 5.1.0 plus the custom `go-training.css` theme.
- `README.md` and `CONTRIBUTING.md` describe how to use the tools.
- The git history is a clean sequence of small, conventional commits — one per task (or per TDD cycle within a task).

The next plan (Plan B) builds the static-site generator (`tools/build-index/`), the deploy `Dockerfile`, the Cloud Run service config, and the GitHub Actions workflows.
