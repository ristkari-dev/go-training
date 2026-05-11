# Theming Overhaul Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current dark+cyan visual design with a Go-themed dark palette (gopher blue + Go green), Source Sans + Source Code Pro typography, a phase-grouped lesson card grid on the landing page, Dracula syntax highlighting, and a vertical text → ↓ code drill-down convention for slides. Apply platform-wide (templates + scaffolder) and retrofit existing lessons 01 and 02.

**Architecture:** Eleven focused commits. CSS design tokens live in both `tools/build-index/index.html.tmpl` (landing page) and `shared/reveal/theme/go-training.css` (slides), so a colour or font change in one place ripples everywhere by editing the token value. Dracula's `highlight.js` theme is vendored alongside the existing reveal.js plugins. The build-index tool gains a static `allLessons` master list so unpublished lessons render as faded placeholders. The slide-authoring convention (vertical `--` separator between prose and code) is encoded in the scaffolder template and applied to existing lessons.

**Tech Stack:** CSS custom properties, reveal.js 5.1.0 (already vendored), highlight.js dracula theme (vendored in this plan), Go 1.23 stdlib (`text/template`, `embed`, `testing`), GNU Make.

---

## Scope

This plan implements every section of the design spec at `docs/superpowers/specs/2026-05-08-theming-overhaul-design.md`:

- Section 1 (palette tokens) → Tasks 2, 6.
- Section 2 (typography) → Tasks 2, 6.
- Section 3 (landing page) → Tasks 5, 6.
- Section 4 (slide theme + Dracula) → Tasks 1, 2, 3, 4.
- Section 5 (slide-authoring convention) → Tasks 7, 8, 9, 10.

Task 11 is end-to-end verification.

---

## File Structure

After the plan completes:

```
shared/reveal/
├── theme/go-training.css                  (rewritten — Task 2)
└── plugin/highlight/dracula.css           (NEW — Task 1, vendored)

tools/
├── build-index/
│   ├── main.go                            (modified — Task 5)
│   ├── main_test.go                       (modified — Task 5)
│   └── index.html.tmpl                    (rewritten — Task 6)
└── new-lesson/template/
    └── slides/
        ├── index.html.tmpl                (modified — Task 3, monokai→dracula)
        └── slides.md.tmpl                 (modified — Task 7, add `--` separators)

CONTRIBUTING.md                            (modified — Task 8, "Code goes down" paragraph)

lessons/01-hello/slides/
├── index.html                              (modified — Task 4, monokai→dracula)
└── slides.md                               (retrofitted — Task 9)

lessons/02-variables/slides/
├── index.html                              (modified — Task 4, monokai→dracula)
└── slides.md                               (retrofitted — Task 10)
```

### Decomposition rationale

- **Dracula CSS comes first (Task 1)** so the new slide theme can reference it without a chicken-and-egg.
- **Theme rewrite (Task 2) before any template updates** so visual changes are isolated to one commit.
- **Template index.html updates (Tasks 3, 4) separate from theme** so a regression in either is bisectable.
- **build-index changes (Tasks 5, 6) split between Go logic and HTML template** because the Go side has tests and the HTML side is content-only.
- **Lesson retrofits (Tasks 9, 10) are last** so the new theme is already serving them by the time their structure changes.

---

## Conventions

- **Working directory:** `/Users/ristkari/code/private/go-training/`
- **Branch:** `feature/theming-overhaul` (already created; spec already committed)
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `style:`)
- **Verification:** every Go-touching task ends with `make test` + `golangci-lint run ./...` passing
- **No new lessons in this plan:** Plans F-K continue separately

---

## Task 1: Vendor Dracula highlight.js theme

Vendor the Dracula CSS file alongside the existing reveal.js plugins so the new slide theme can swap from Monokai. Dracula is part of highlight.js's standard theme set but not in reveal.js's bundled plugins.

**Files:**
- Create: `shared/reveal/plugin/highlight/dracula.css`

- [ ] **Step 1: Confirm you're on the feature branch**

```bash
cd /Users/ristkari/code/private/go-training
git status -sb
git log --oneline main..HEAD
```

Expected: `## feature/theming-overhaul`, exactly one commit ahead of main ("Add theming overhaul design spec"). If you see something else, STOP and report — do not recreate the branch.

- [ ] **Step 2: Download Dracula CSS from highlight.js's GitHub release**

Use highlight.js v11.10.0's `dracula.css` from the official styles directory:

```bash
curl -fsSL -o shared/reveal/plugin/highlight/dracula.css \
  https://raw.githubusercontent.com/highlightjs/highlight.js/11.10.0/src/styles/dracula.css
```

If `curl` is unavailable, an equivalent `wget` command:

```bash
wget -O shared/reveal/plugin/highlight/dracula.css \
  https://raw.githubusercontent.com/highlightjs/highlight.js/11.10.0/src/styles/dracula.css
```

- [ ] **Step 3: Verify the file is present and looks like Dracula**

```bash
test -f shared/reveal/plugin/highlight/dracula.css && echo "exists"
wc -l shared/reveal/plugin/highlight/dracula.css
grep -c "#282a36" shared/reveal/plugin/highlight/dracula.css
grep -c "#ff79c6" shared/reveal/plugin/highlight/dracula.css
```

Expected:
- `exists`
- ~60-80 lines (the file is small)
- `#282a36` appears at least once (Dracula's charcoal background)
- `#ff79c6` appears at least once (Dracula's pink keyword)

If counts are zero, the download fetched the wrong file — STOP and report BLOCKED.

- [ ] **Step 4: Commit**

```bash
git add shared/reveal/plugin/highlight/dracula.css
git commit -m "build: vendor highlight.js Dracula theme"
```

---

## Task 2: Rewrite the slide theme

Replace `shared/reveal/theme/go-training.css` with the new tokens, typography, and slide-pattern classes.

**Files:**
- Modify: `shared/reveal/theme/go-training.css` (full replacement)

- [ ] **Step 1: Replace `shared/reveal/theme/go-training.css`** with:

```css
/*
 * go-training reveal.js theme — Go-themed dark
 *
 * Design tokens, type scale, slide-pattern classes, and reveal.js
 * variable overrides. See docs/superpowers/specs/2026-05-08-theming-overhaul-design.md.
 */

@import url("../dist/theme/fonts/source-sans-pro/source-sans-pro.css");

:root {
  /* Background and surface */
  --bg: #1d2541;
  --surface: #252e4f;
  --surface-2: #2c3658;
  --border: rgba(154, 230, 180, 0.18);

  /* Foreground */
  --fg: #e8eef9;
  --fg-muted: rgba(232, 238, 249, 0.65);
  --fg-subtle: rgba(232, 238, 249, 0.45);

  /* Accent (Go green) */
  --accent: #9ae6b4;
  --accent-soft: rgba(154, 230, 180, 0.10);
  --accent-strong: rgba(154, 230, 180, 0.45);

  /* Code (Dracula) */
  --code-bg: #282a36;

  /* Error (Common mistake error blocks) */
  --error-accent: #ff79c6;
  --error-bg: rgba(255, 121, 198, 0.10);

  /* Fonts */
  --font-sans: "Source Sans 3", "Source Sans Pro", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-mono: "Source Code Pro", "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
}

/* Reveal.js variable overrides — these cascade into reveal.js's defaults */
.reveal {
  --r-background-color: var(--bg);
  --r-main-color: var(--fg);
  --r-main-font: var(--font-sans);
  --r-main-font-size: 38px;
  --r-heading-color: var(--fg);
  --r-heading-font: var(--font-sans);
  --r-heading-text-shadow: none;
  --r-heading-text-transform: none;
  --r-heading1-size: 2.2em;
  --r-heading2-size: 1.4em;
  --r-heading3-size: 1.0em;
  --r-link-color: var(--accent);
  --r-link-color-hover: var(--fg);
  --r-selection-color: var(--bg);
  --r-selection-background-color: var(--accent);
  --r-code-font: var(--font-mono);

  background-color: var(--bg);
  color: var(--fg);
}

/* Headings */
.reveal h1 {
  font-weight: 700;
  letter-spacing: -0.025em;
  line-height: 1.05;
  margin-bottom: 0.4em;
}
.reveal h2 {
  font-weight: 600;
  letter-spacing: -0.01em;
  line-height: 1.15;
  margin-bottom: 0.5em;
}
.reveal h3 {
  font-family: var(--font-mono);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.10em;
  color: var(--accent);
  margin-bottom: 0.6em;
}
.reveal h4 {
  font-weight: 600;
  letter-spacing: -0.005em;
}

/* Body text */
.reveal p {
  color: rgba(232, 238, 249, 0.85);
  line-height: 1.55;
}

/* Inline code */
.reveal code {
  font-family: var(--font-mono);
  color: var(--accent);
  background: var(--accent-soft);
  padding: 0.05em 0.4em;
  border-radius: 3px;
  font-size: 0.95em;
}

/* Code blocks */
.reveal pre {
  width: 100%;
  font-size: 0.55em;
  box-shadow: none;
  border-radius: 8px;
  background: var(--code-bg);
  padding: 0;
  margin: 0.5em 0;
}
.reveal pre code {
  padding: 1em 1.2em;
  max-height: 75vh;
  overflow: auto;
  line-height: 1.55;
  background: transparent;
  color: #f8f8f2;
  font-family: var(--font-mono);
  font-size: inherit;
  border-radius: 8px;
  /* highlight.js dracula.css supplies token colors */
}

/* Lists */
.reveal ul {
  display: block;
  margin-left: 0;
  padding-left: 0;
  list-style: none;
}
.reveal ul li {
  position: relative;
  padding-left: 1.25em;
  margin: 0.35em 0;
  color: rgba(232, 238, 249, 0.88);
}
.reveal ul li::before {
  content: "";
  position: absolute;
  left: 0.3em;
  top: 0.6em;
  width: 0.35em;
  height: 0.35em;
  border-radius: 50%;
  background: var(--accent);
}
.reveal ol {
  color: rgba(232, 238, 249, 0.88);
}
.reveal ol li::marker {
  color: var(--accent);
  font-weight: 600;
}

/* Blockquote — used as the default callout */
.reveal blockquote {
  background: var(--accent-soft);
  border-left: 4px solid var(--accent);
  padding: 0.6em 1em;
  border-radius: 0 8px 8px 0;
  font-style: normal;
  width: 100%;
  box-shadow: none;
}

/* Images */
.reveal section img {
  border: none;
  background: transparent;
  box-shadow: none;
}

/* Slide number indicator */
.reveal .slide-number {
  font-family: var(--font-mono);
  background: transparent;
  color: var(--fg-subtle);
  font-size: 0.6em;
}

/* Progress bar */
.reveal .progress {
  color: var(--accent);
}

/* Navigation arrows */
.reveal .controls {
  color: var(--accent);
}

/* Title-slide lesson label — e.g. "LESSON 01 · PHASE 1 — FOUNDATIONS" */
.reveal .lesson-label {
  font-family: var(--font-mono);
  font-size: 0.6em;
  color: var(--accent);
  letter-spacing: 0.18em;
  text-transform: uppercase;
  font-weight: 600;
  margin-bottom: 0.6em;
}

/* Callout — boxed accent block (Learning goal, Note: blocks) */
.reveal .callout {
  background: var(--accent-soft);
  border-left: 4px solid var(--accent);
  padding: 0.8em 1.2em;
  border-radius: 0 8px 8px 0;
  margin: 0.6em 0;
}
.reveal .callout h3 {
  margin-bottom: 0.3em;
}

/* Error block — pink-accent code-styled block for compiler errors / unexpected output */
.reveal .error-block,
.reveal pre.error-block {
  font-family: var(--font-mono);
  background: var(--error-bg);
  color: #f8f8f2;
  border-left: 4px solid var(--error-accent);
  padding: 0.7em 1em;
  border-radius: 0 6px 6px 0;
  font-size: 0.55em;
  line-height: 1.55;
  margin: 0.4em 0;
}
.reveal .error-block code {
  background: transparent;
  color: inherit;
  padding: 0;
}
```

- [ ] **Step 2: Verify the file is well-formed CSS**

```bash
# Confirm length and a couple of unique markers
wc -l shared/reveal/theme/go-training.css
grep -c "^:root" shared/reveal/theme/go-training.css     # 1
grep -c "^.reveal" shared/reveal/theme/go-training.css   # at least 15
grep -c "9ae6b4" shared/reveal/theme/go-training.css     # appears 5+ times
```

Expected: ~190-220 lines, one `:root`, many `.reveal` selectors, several `#9ae6b4` references.

- [ ] **Step 3: Commit**

```bash
git add shared/reveal/theme/go-training.css
git commit -m "style(theme): rewrite slide theme with Go-themed tokens + Dracula code"
```

---

## Task 3: Update the lesson template's `index.html.tmpl`

Swap the highlight stylesheet from `monokai.css` to `dracula.css` so new lessons get Dracula by default.

**Files:**
- Modify: `tools/new-lesson/template/slides/index.html.tmpl`

- [ ] **Step 1: Edit the highlight stylesheet line**

Open `tools/new-lesson/template/slides/index.html.tmpl`. Find:

```html
  <link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/monokai.css">
```

Replace with:

```html
  <link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/dracula.css">
```

Nothing else in the template changes.

- [ ] **Step 2: Verify the swap**

```bash
grep -c "dracula.css" tools/new-lesson/template/slides/index.html.tmpl   # 1
grep -c "monokai.css" tools/new-lesson/template/slides/index.html.tmpl   # 0
```

- [ ] **Step 3: Confirm scaffolder tests still pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: all 6 test functions still pass. The scaffolder tests don't check the content of `index.html.tmpl`, only that it exists and is copied — but a syntax error in the file would break the embed, so this catches that.

- [ ] **Step 4: Smoke-test by scaffolding a sandbox lesson**

```bash
TMP=$(mktemp -d)
go run ./tools/new-lesson -name 99-demo -out "$TMP"
grep -c "dracula.css" "$TMP/99-demo/slides/index.html"   # 1
rm -rf "$TMP"
```

Expected: `1`.

- [ ] **Step 5: Commit**

```bash
git add tools/new-lesson/template/slides/index.html.tmpl
git commit -m "style(new-lesson): swap monokai for dracula in lesson template"
```

---

## Task 4: Update existing lessons' `index.html` to use Dracula

Lessons 01 and 02's `slides/index.html` were generated from the template before Task 3 — they still reference `monokai.css`. Bring them in line with the new template.

**Files:**
- Modify: `lessons/01-hello/slides/index.html`
- Modify: `lessons/02-variables/slides/index.html`

- [ ] **Step 1: Update lesson 01**

In `lessons/01-hello/slides/index.html`, find:

```html
  <link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/monokai.css">
```

Replace with:

```html
  <link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/dracula.css">
```

- [ ] **Step 2: Update lesson 02**

Same edit in `lessons/02-variables/slides/index.html`.

- [ ] **Step 3: Verify both lessons swapped**

```bash
grep -c "dracula.css" lessons/01-hello/slides/index.html       # 1
grep -c "monokai.css" lessons/01-hello/slides/index.html       # 0
grep -c "dracula.css" lessons/02-variables/slides/index.html   # 1
grep -c "monokai.css" lessons/02-variables/slides/index.html   # 0
```

- [ ] **Step 4: Smoke-test by serving a deck locally**

```bash
make slides-dev LESSON=01-hello &
SDPID=$!
sleep 2
curl -sS http://localhost:8000/lessons/01-hello/slides/index.html | grep -c "dracula.css"
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/shared/reveal/plugin/highlight/dracula.css
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `1`, then `200` (Dracula CSS is served from the vendored location).

- [ ] **Step 5: Commit**

```bash
git add lessons/01-hello/slides/index.html lessons/02-variables/slides/index.html
git commit -m "style(lessons): swap monokai for dracula in lessons 01-02 index.html"
```

---

## Task 5: Add `allLessons` + Phase logic to build-index (TDD)

Restructure `tools/build-index/main.go` so the landing page renders all 29 lessons grouped by phase, with placeholder cards for unpublished ones. Drops the README-based `extractTitle` flow (titles now come from a static master list).

**Files:**
- Modify: `tools/build-index/main.go`
- Modify: `tools/build-index/main_test.go`

- [ ] **Step 1: Replace `tools/build-index/main.go`** with:

```go
// Command build-index produces a static slides site under dist/: every lesson's
// slides copied into place, the shared reveal.js assets copied alongside, and a
// generated index.html landing page listing all lessons grouped by phase.
package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
)

//go:embed index.html.tmpl
var indexFS embed.FS

// lessonInfo is one entry in the master lesson list. The master list is
// authoritative for titles and blurbs; whether a lesson is published is
// determined at build time by checking whether the lesson's directory exists
// on disk.
type lessonInfo struct {
	Number string // "01"
	Slug   string // "hello"
	Title  string // "Hello, Go"
	Blurb  string // "go run · package main · fmt"
	Phase  int    // 1-4
}

// allLessons is the master list of all 29 lessons across the four phases.
// New lessons are added here when they're planned (so they appear as faded
// placeholders on the landing page until their content lands on disk).
var allLessons = []lessonInfo{
	// Phase 1 — Foundations
	{Number: "01", Slug: "hello", Title: "Hello, Go", Blurb: "go run · package main · fmt", Phase: 1},
	{Number: "02", Slug: "variables", Title: "Variables, types, operators", Blurb: "var · := · float64 · const", Phase: 1},
	{Number: "03", Slug: "control-flow", Title: "Control flow", Blurb: "if · for · switch", Phase: 1},
	{Number: "04", Slug: "functions", Title: "Functions & first tests", Blurb: "multi-return · table tests", Phase: 1},
	{Number: "05", Slug: "slices-maps", Title: "Slices and maps", Blurb: "[]T · map[K]V · range", Phase: 1},
	{Number: "06", Slug: "structs", Title: "Structs & methods", Blurb: "type T struct · methods", Phase: 1},
	{Number: "07", Slug: "packages", Title: "Packages & modules", Blurb: "go.mod · imports · gofmt", Phase: 1},
	{Number: "08", Slug: "capstone", Title: "Phase 1 capstone", Blurb: "expense tracker CLI", Phase: 1},
	// Phase 2 — Idiomatic Go
	{Number: "09", Slug: "pointers", Title: "Pointers", Blurb: "value vs reference", Phase: 2},
	{Number: "10", Slug: "interfaces", Title: "Interfaces", Blurb: "io.Reader · any", Phase: 2},
	{Number: "11", Slug: "errors", Title: "Errors", Blurb: "wrapping · errors.Is/As", Phase: 2},
	{Number: "12", Slug: "generics", Title: "Generics", Blurb: "type parameters", Phase: 2},
	{Number: "13", Slug: "encoding-io", Title: "Encoding & I/O", Blurb: "JSON · bufio · streams", Phase: 2},
	{Number: "14", Slug: "stdlib", Title: "Time, strings, regex", Blurb: "stdlib literacy", Phase: 2},
	{Number: "15", Slug: "structure", Title: "Project structure", Blurb: "cmd/ · internal/", Phase: 2},
	// Phase 3 — Concurrency & Systems
	{Number: "16", Slug: "goroutines", Title: "Goroutines & channels", Blurb: "go · chan · range", Phase: 3},
	{Number: "17", Slug: "select", Title: "Select & timers", Blurb: "select · time.After", Phase: 3},
	{Number: "18", Slug: "sync", Title: "sync & memory model", Blurb: "Mutex · race detector", Phase: 3},
	{Number: "19", Slug: "context", Title: "context", Blurb: "cancellation · deadlines", Phase: 3},
	{Number: "20", Slug: "patterns", Title: "Concurrency patterns", Blurb: "worker pool · errgroup", Phase: 3},
	{Number: "21", Slug: "networking", Title: "Networking", Blurb: "net · TCP · syscalls", Phase: 3},
	{Number: "22", Slug: "profiling", Title: "Profiling & benchmarks", Blurb: "pprof · go test -bench", Phase: 3},
	// Phase 4 — Production & Distributed
	{Number: "23", Slug: "http-server", Title: "HTTP servers", Blurb: "net/http · slog", Phase: 4},
	{Number: "24", Slug: "http-client", Title: "HTTP clients", Blurb: "retries · timeouts", Phase: 4},
	{Number: "25", Slug: "grpc", Title: "gRPC", Blurb: "protobuf · streaming", Phase: 4},
	{Number: "26", Slug: "config", Title: "Config & shutdown", Blurb: "flags · env · signals", Phase: 4},
	{Number: "27", Slug: "container", Title: "Containerization", Blurb: "multi-stage · distroless", Phase: 4},
	{Number: "28", Slug: "observability", Title: "Observability", Blurb: "metrics · traces", Phase: 4},
	{Number: "29", Slug: "capstone-final", Title: "Course capstone", Blurb: "distributed wrap-up", Phase: 4},
}

// phaseInfo names the four phases for rendering.
type phaseInfo struct {
	Number int
	Name   string
}

var phases = []phaseInfo{
	{Number: 1, Name: "Foundations"},
	{Number: 2, Name: "Idiomatic Go"},
	{Number: 3, Name: "Concurrency & Systems"},
	{Number: 4, Name: "Production & Distributed"},
}

// lessonView is what the template iterates over.
type lessonView struct {
	Number    string
	Title     string
	Blurb     string
	Name      string // "01-hello" — used for slides/ URL when Published
	Published bool
}

// phaseView is a phase plus its lessons.
type phaseView struct {
	Number  int
	Name    string
	Lessons []lessonView
}

// indexData is the root template payload.
type indexData struct {
	Phases []phaseView
}

// collectPublishedSlugs walks lessonsDir to find which "NN-slug" directories
// exist on disk. Returns a set keyed by full lesson name (e.g. "01-hello").
func collectPublishedSlugs(lessonsDir string) (map[string]bool, error) {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	out := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Confirm a slides/ subdir exists; otherwise it isn't really published.
		if _, err := os.Stat(filepath.Join(lessonsDir, e.Name(), "slides")); err != nil {
			continue
		}
		out[e.Name()] = true
	}
	return out, nil
}

// buildPhases returns the four phases populated with lessons from allLessons,
// each lesson marked Published if its directory exists in lessonsDir.
func buildPhases(lessonsDir string) ([]phaseView, error) {
	published, err := collectPublishedSlugs(lessonsDir)
	if err != nil {
		return nil, err
	}
	byPhase := map[int][]lessonView{}
	for _, l := range allLessons {
		name := l.Number + "-" + l.Slug
		byPhase[l.Phase] = append(byPhase[l.Phase], lessonView{
			Number:    l.Number,
			Title:     l.Title,
			Blurb:     l.Blurb,
			Name:      name,
			Published: published[name],
		})
	}
	// Sort each phase's lessons by Number (string sort works because they're
	// all two-digit zero-padded).
	for k := range byPhase {
		sort.Slice(byPhase[k], func(i, j int) bool {
			return byPhase[k][i].Number < byPhase[k][j].Number
		})
	}
	out := make([]phaseView, 0, len(phases))
	for _, p := range phases {
		out = append(out, phaseView{
			Number:  p.Number,
			Name:    p.Name,
			Lessons: byPhase[p.Number],
		})
	}
	return out, nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close() //nolint:errcheck
		return err
	}
	return out.Close()
}

func renderIndex(data indexData) ([]byte, error) {
	tmpl, err := template.ParseFS(indexFS, "index.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

func build(lessonsDir, sharedDir, outDir string) error {
	if err := os.RemoveAll(outDir); err != nil {
		return fmt.Errorf("clear out dir: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	phases, err := buildPhases(lessonsDir)
	if err != nil {
		return fmt.Errorf("build phases: %w", err)
	}
	// Copy every published lesson's slides.
	for _, p := range phases {
		for _, l := range p.Lessons {
			if !l.Published {
				continue
			}
			src := filepath.Join(lessonsDir, l.Name, "slides")
			dst := filepath.Join(outDir, "lessons", l.Name, "slides")
			if err := copyTree(src, dst); err != nil {
				return fmt.Errorf("copy %s slides: %w", l.Name, err)
			}
		}
	}
	// Copy shared/reveal.
	if _, err := os.Stat(sharedDir); err == nil {
		if err := copyTree(sharedDir, filepath.Join(outDir, "shared", "reveal")); err != nil {
			return fmt.Errorf("copy shared/reveal: %w", err)
		}
	}
	idx, err := renderIndex(indexData{Phases: phases})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "index.html"), idx, 0o644)
}

func main() {
	os.Exit(runWithArgs(os.Args))
}

func runWithArgs(args []string) int {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	lessonsDir := flags.String("lessons", "lessons", "directory containing lesson folders")
	sharedDir := flags.String("shared", "shared/reveal", "directory containing shared reveal.js assets")
	outDir := flags.String("out", "dist", "output directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if err := build(*lessonsDir, *sharedDir, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	fmt.Printf("built %s\n", *outDir)
	return 0
}
```

- [ ] **Step 2: Replace `tools/build-index/main_test.go`** with:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAllLessonsCovers29Lessons(t *testing.T) {
	if len(allLessons) != 29 {
		t.Errorf("allLessons has %d entries, want 29", len(allLessons))
	}
	seen := map[string]bool{}
	for _, l := range allLessons {
		if seen[l.Number] {
			t.Errorf("duplicate lesson number %q", l.Number)
		}
		seen[l.Number] = true
		if l.Phase < 1 || l.Phase > 4 {
			t.Errorf("lesson %s: phase %d outside 1-4", l.Number, l.Phase)
		}
		if l.Title == "" || l.Slug == "" || l.Blurb == "" {
			t.Errorf("lesson %s: empty Title/Slug/Blurb", l.Number)
		}
	}
}

func TestCollectPublishedSlugs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "01-hello", "slides", "slides.md"), "# hi")
	writeFile(t, filepath.Join(dir, "02-variables", "slides", "index.html"), "<html></html>")
	// A dir without slides/ — should be ignored.
	if err := os.MkdirAll(filepath.Join(dir, "03-broken"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := collectPublishedSlugs(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"01-hello": true, "02-variables": true}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing %q", k)
		}
	}
	if got["03-broken"] {
		t.Errorf("03-broken should not appear (no slides/ subdir)")
	}
}

func TestCollectPublishedSlugsHandlesMissingDir(t *testing.T) {
	got, err := collectPublishedSlugs(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d, want 0", len(got))
	}
}

func TestBuildPhases(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "01-hello", "slides", "slides.md"), "# hi")
	writeFile(t, filepath.Join(dir, "02-variables", "slides", "slides.md"), "# hi")

	phases, err := buildPhases(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(phases) != 4 {
		t.Fatalf("got %d phases, want 4", len(phases))
	}
	// Phase 1 should have 8 lessons.
	if len(phases[0].Lessons) != 8 {
		t.Errorf("phase 1 has %d lessons, want 8", len(phases[0].Lessons))
	}
	// Phase 1 lessons should be sorted by Number.
	for i, want := range []string{"01", "02", "03", "04", "05", "06", "07", "08"} {
		if phases[0].Lessons[i].Number != want {
			t.Errorf("phase 1 lesson %d: got %q, want %q", i, phases[0].Lessons[i].Number, want)
		}
	}
	// Published flags.
	if !phases[0].Lessons[0].Published {
		t.Error("01-hello should be Published")
	}
	if !phases[0].Lessons[1].Published {
		t.Error("02-variables should be Published")
	}
	if phases[0].Lessons[2].Published {
		t.Error("03 should NOT be Published")
	}
}

func TestRenderIndexIncludesAllPhases(t *testing.T) {
	phases, err := buildPhases(t.TempDir()) // empty dir; everything unpublished
	if err != nil {
		t.Fatal(err)
	}
	body, err := renderIndex(indexData{Phases: phases})
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"Foundations", "Idiomatic Go", "Concurrency", "Production"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing phase %q", want)
		}
	}
	// Lesson titles from allLessons appear.
	for _, want := range []string{"Hello, Go", "Variables, types, operators", "Goroutines"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing lesson %q", want)
		}
	}
}

func TestCopyTreeMirrorsStructure(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	writeFile(t, filepath.Join(src, "a", "b", "c.txt"), "hello")
	writeFile(t, filepath.Join(src, "a", "d.txt"), "world")
	if err := copyTree(src, dst); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dst, "a", "b", "c.txt"))
	if err != nil || string(body) != "hello" {
		t.Errorf("a/b/c.txt: got %q, err %v", body, err)
	}
}

func TestBuildEndToEnd(t *testing.T) {
	root := t.TempDir()
	lessonsDir := filepath.Join(root, "lessons")
	sharedDir := filepath.Join(root, "shared", "reveal")
	outDir := filepath.Join(root, "dist")

	writeFile(t, filepath.Join(lessonsDir, "01-hello", "slides", "index.html"), "<html>hi</html>")
	writeFile(t, filepath.Join(lessonsDir, "01-hello", "slides", "slides.md"), "# Hello\n")
	writeFile(t, filepath.Join(sharedDir, "dist", "reveal.js"), "console.log('reveal');")

	if err := build(lessonsDir, sharedDir, outDir); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		"index.html",
		"lessons/01-hello/slides/index.html",
		"lessons/01-hello/slides/slides.md",
		"shared/reveal/dist/reveal.js",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	idx, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(idx), "Hello, Go") {
		t.Errorf("index.html missing 'Hello, Go'\n%s", idx)
	}
}
```

- [ ] **Step 3: Run the tests against the existing `index.html.tmpl`**

```bash
go test ./tools/build-index/ -v
```

Expected: most tests pass; `TestRenderIndexIncludesAllPhases` and `TestRenderIndexEmpty`-style tests may FAIL because the OLD `index.html.tmpl` doesn't iterate over `Phases`. That's expected — Task 6 replaces the template.

If ALL tests fail, the Go code has a bug — STOP and report.

- [ ] **Step 4: Run lint**

```bash
golangci-lint run ./tools/build-index/...
```

Expected: 0 issues. If `errcheck` flags the `Close()` calls in `copyFile`, the `//nolint:errcheck` comments should already be in place.

- [ ] **Step 5: Commit**

```bash
git add tools/build-index/main.go tools/build-index/main_test.go
git commit -m "feat(build-index): add allLessons + buildPhases for phase-grouped landing"
```

(Note: the template still needs to be updated in Task 6 to match the new `indexData` shape. Until Task 6, `make slides-build` produces broken HTML output. That's expected.)

---

## Task 6: Rewrite the landing-page template

Replace `tools/build-index/index.html.tmpl` with the new phase-grouped grid.

**Files:**
- Modify: `tools/build-index/index.html.tmpl` (full replacement)

- [ ] **Step 1: Replace `tools/build-index/index.html.tmpl`** with:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Go Training</title>
  <style>
    @import url("shared/reveal/dist/theme/fonts/source-sans-pro/source-sans-pro.css");

    :root {
      --bg: #1d2541;
      --surface: #252e4f;
      --surface-2: #2c3658;
      --border: rgba(154, 230, 180, 0.18);

      --fg: #e8eef9;
      --fg-muted: rgba(232, 238, 249, 0.65);
      --fg-subtle: rgba(232, 238, 249, 0.45);

      --accent: #9ae6b4;
      --accent-soft: rgba(154, 230, 180, 0.10);
      --accent-strong: rgba(154, 230, 180, 0.45);

      --font-sans: "Source Sans 3", "Source Sans Pro", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      --font-mono: "Source Code Pro", "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
    }

    * { box-sizing: border-box; }
    html, body { margin: 0; padding: 0; background: var(--bg); color: var(--fg); }
    body { font-family: var(--font-sans); font-size: 16px; line-height: 1.55; }

    main { max-width: 960px; margin: 0 auto; padding: 60px 28px 100px; }

    h1 {
      font-size: 2.6rem;
      font-weight: 700;
      color: var(--accent);
      letter-spacing: -0.025em;
      margin: 0 0 0.3rem;
    }
    p.lead {
      font-size: 1.1rem;
      color: var(--fg-muted);
      margin: 0 0 2.5rem;
      max-width: 640px;
    }

    .phase {
      font-family: var(--font-mono);
      font-size: 0.75rem;
      font-weight: 600;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--accent);
      margin: 2.5rem 0 1rem;
      display: flex;
      align-items: center;
      gap: 0.7rem;
    }
    .phase::after {
      content: "";
      flex: 1;
      height: 1px;
      background: var(--border);
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
      gap: 0.7rem;
    }

    .lesson {
      background: var(--accent-soft);
      border: 1px solid var(--border);
      border-radius: 10px;
      padding: 0.95rem 1rem 1rem;
      text-decoration: none;
      color: inherit;
      transition: transform 150ms ease, border-color 150ms ease, background 150ms ease;
      display: block;
    }
    .lesson:hover {
      transform: translateY(-2px);
      border-color: var(--accent-strong);
      background: rgba(154, 230, 180, 0.13);
    }
    .lesson:focus-visible {
      outline: 2px solid var(--accent);
      outline-offset: 3px;
    }
    .lesson .num {
      font-family: var(--font-mono);
      font-size: 0.7rem;
      font-weight: 600;
      color: var(--accent);
      letter-spacing: 0.05em;
    }
    .lesson .title {
      font-size: 1.05rem;
      font-weight: 600;
      color: var(--fg);
      letter-spacing: -0.01em;
      margin-top: 0.25rem;
      line-height: 1.25;
    }
    .lesson .blurb {
      font-family: var(--font-mono);
      font-size: 0.78rem;
      color: var(--fg-muted);
      margin-top: 0.45rem;
      line-height: 1.4;
    }

    .lesson.future {
      background: rgba(255, 255, 255, 0.02);
      border: 1px dashed rgba(255, 255, 255, 0.10);
      opacity: 0.42;
      cursor: default;
    }
    .lesson.future:hover {
      transform: none;
      background: rgba(255, 255, 255, 0.02);
      border-color: rgba(255, 255, 255, 0.10);
    }
    .lesson.future .num,
    .lesson.future .title { color: var(--fg-muted); }
    .lesson.future .blurb { color: var(--fg-subtle); }

    footer {
      margin-top: 4rem;
      padding-top: 1.5rem;
      border-top: 1px solid rgba(255,255,255,0.06);
      font-size: 0.9rem;
      color: var(--fg-subtle);
    }
    footer a {
      color: var(--accent);
      text-decoration: none;
    }
    footer a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <main>
    <h1>Go Training</h1>
    <p class="lead">A Go programming course delivered as code + per-lesson reveal.js slide decks. Starts at programming-101 and finishes with concurrency, systems programming, production services, and distributed patterns.</p>

    {{range .Phases}}
    <div class="phase">Phase {{.Number}} · {{.Name}}</div>
    <div class="grid">
      {{range .Lessons}}
      {{if .Published}}
      <a class="lesson" href="lessons/{{.Name}}/slides/">
        <div class="num">{{.Number}}</div>
        <div class="title">{{.Title}}</div>
        <div class="blurb">{{.Blurb}}</div>
      </a>
      {{else}}
      <div class="lesson future" aria-disabled="true">
        <div class="num">{{.Number}}</div>
        <div class="title">{{.Title}}</div>
        <div class="blurb">{{.Blurb}}</div>
      </div>
      {{end}}
      {{end}}
    </div>
    {{end}}

    <footer>
      Source: <a href="https://github.com/ristkari-dev/go-training">github.com/ristkari-dev/go-training</a>
    </footer>
  </main>
</body>
</html>
```

- [ ] **Step 2: Run the tests — they should all pass now**

```bash
go test ./tools/build-index/ -v
```

Expected: every test passes — including `TestRenderIndexIncludesAllPhases` and `TestBuildEndToEnd`.

- [ ] **Step 3: Verify the rendered output locally**

```bash
make slides-build
grep -c "^    <div class=\"phase\">Phase " dist/index.html         # 4 (one per phase)
grep -c "Hello, Go" dist/index.html                                 # at least 1
grep -c "class=\"lesson future\"" dist/index.html                  # 27 (29 lessons - 2 published)
rm -rf dist
```

Expected counts: 4, ≥1, 27.

- [ ] **Step 4: Run lint**

```bash
golangci-lint run ./...
```

Expected: 0 issues.

- [ ] **Step 5: Commit**

```bash
git add tools/build-index/index.html.tmpl
git commit -m "style(build-index): rewrite landing as phase-grouped lesson card grid"
```

---

## Task 7: Update the scaffolder slides template

Add `--` separators in the pre-stocked code-bearing sub-slides of `tools/new-lesson/template/slides/slides.md.tmpl`, so new lessons start with the right vertical-stack structure.

**Files:**
- Modify: `tools/new-lesson/template/slides/slides.md.tmpl`

- [ ] **Step 1: Replace `tools/new-lesson/template/slides/slides.md.tmpl`** with:

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

TODO — short intro to the concept, what we'll show in the code below.

--

### Code

```go
package main

import "fmt"

func main() {
	fmt.Println("TODO: minimal example introducing the concept")
}
```

---

### A worked example

TODO — short intro to a substantive example.

--

### Code

```go
package main

func main() {
	// TODO
}
```

---

### Common mistake

TODO — what NOT to do, in prose. The broken code is on the next slide down.

--

### Code (broken) and what Go says

```go
// the wrong way
```

TODO — fix and explanation.

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

TODO — short intro.

--

### Code

```go
// minimal
```

---

### A worked example

TODO

--

### Code

```go
// concrete
```

---

### Common mistake

TODO

--

### Code (broken)

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

TODO

--

### Code

```go
// minimal
```

---

### A worked example

TODO

--

### Code

```go
// concrete
```

---

### Common mistake

TODO

--

### Code (broken)

```go
// wrong way
```

---

### Recap

- TODO

---

## Practice

### Warm-up

TODO — short description. Files: `exercises/warmup.go`, `exercises/warmup_test.go`.

```bash
cd lessons/{{.Name}}/exercises
go test -run Warmup -v
```

---

### Main

TODO — description. Files: `exercises/main.go`, `exercises/main_test.go`.

```bash
cd lessons/{{.Name}}/exercises
go test -v
```

Note:
Speaker notes for the live lecture go here. They render only in the speaker view (press `S`).

---

## What we learned

- TODO
- TODO

---

## Up next

Lesson NN — TODO
````

> Reminder: when writing the actual file, use only three-backtick fences (the outer four-backtick wrapper is for this doc).

- [ ] **Step 2: Verify the structure with greps**

```bash
grep -c "^--$" tools/new-lesson/template/slides/slides.md.tmpl    # 9 (3 concepts × 3 code-bearing sub-slides)
grep -c "^---$" tools/new-lesson/template/slides/slides.md.tmpl   # should still be reasonable (~25)
grep -c "^## Concept " tools/new-lesson/template/slides/slides.md.tmpl  # 3
```

Expected: 9 `--` separators (one per code-bearing sub-slide across 3 concepts × 3 each), ~25 `---` separators, 3 Concept blocks.

- [ ] **Step 3: Confirm scaffolder tests still pass**

```bash
go test ./tools/new-lesson/ -v
```

Expected: every test passes.

- [ ] **Step 4: Smoke-test a scaffolded lesson renders correctly**

```bash
TMP=$(mktemp -d)
go run ./tools/new-lesson -name 99-demo -out "$TMP"
grep -c "^--$" "$TMP/99-demo/slides/slides.md"   # 9
rm -rf "$TMP"
```

Expected: `9`.

- [ ] **Step 5: Commit**

```bash
git add tools/new-lesson/template/slides/slides.md.tmpl
git commit -m "feat(new-lesson): pre-stock vertical `--` separators in slide template"
```

---

## Task 8: Update CONTRIBUTING with "Code goes down"

Add a paragraph to the Phase 1 conventions section so authors know the rule.

**Files:**
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Find the right insertion point**

```bash
grep -n "^### " CONTRIBUTING.md
```

Find `### Heavy-explanatory slide style`. The new paragraph goes at the END of that subsection (before the next `### …` heading).

- [ ] **Step 2: Append the "Code goes down" paragraph**

Use the Edit tool. The anchor: find the last line of the "Heavy-explanatory slide style" subsection (after the existing 5-bullet pattern list ending with "5. **Recap** — a bullet list of takeaways.") and before the `### When students start writing tests` heading. Insert:

```markdown

**Code goes "down."** When a sub-slide has both explanatory prose and a code block, split them into a vertical stack with `--`. The prose is the parent slide; the code is the child below it. Authors give each vertical step its own `###` sub-heading (e.g. `### Code`, `### Run it`, `### Output`). Students press Right to move between concepts; Down to drill into code. Reveal.js shows a navigation arrow at the bottom-right when there's more below — no need for an explicit "press Down" cue.
```

(Leading blank line + paragraph + trailing blank line.)

- [ ] **Step 3: Verify the addition**

```bash
grep -c "^\*\*Code goes \"down\.\"\*\*" CONTRIBUTING.md   # 1
grep -B1 "^### When students start writing tests$" CONTRIBUTING.md
```

Expected: count `1`, and the line above the "When students start writing tests" header is blank (the inserted blank line).

- [ ] **Step 4: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "docs(contributing): document 'code goes down' slide-stack convention"
```

---

## Task 9: Retrofit lesson 01 slides.md

Apply the vertical-stack convention to every code-bearing sub-slide in `lessons/01-hello/slides/slides.md`.

**Files:**
- Modify: `lessons/01-hello/slides/slides.md`

### The transformation rule

For each `### Sub-slide` that contains a code block:

1. The sub-slide's **introductory prose** (one or two short sentences setting up what's coming) stays on the parent slide.
2. Insert `--` (on its own line, surrounded by blank lines).
3. The **code, output, and any post-code explanation** lives on the child slide, prefixed with a `### Code` (or `### Code (broken)` for common-mistake slides) sub-heading.

If the original sub-slide had more than one logical step inside (e.g. "Save this as hello.go" then "Then run:" then "You should see:"), consolidate those into the single child slide — keep all the code/command/output blocks together.

### Worked example — Concept 1 "The basics" (before / after)

**Before** (slides.md lines 29-55):

```markdown
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
```

**After:**

```markdown
### The basics

Save the following as `hello.go`, then run it with `go run hello.go`. You should see `Hello, Go!` printed — you just compiled and ran a Go program.

--

### Code

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

```bash
go run hello.go
```

Output:

```
Hello, Go!
```
```

### Where the rule applies (lesson 01)

Eleven code-bearing sub-slides need the split:

| Concept | Sub-slide |
|---|---|
| 1. From Go installed to running code | The basics |
| 1. From Go installed to running code | A worked example |
| 1. From Go installed to running code | Common mistake |
| 2. Modules and the project structure | The basics |
| 2. Modules and the project structure | A worked example |
| 3. The main package and main function | The basics |
| 3. The main package and main function | A worked example |
| 3. The main package and main function | Common mistake |
| 4. fmt and basic printing | The basics |
| 4. fmt and basic printing | A worked example |
| 4. fmt and basic printing | Common mistake |

**Not split** (no code in them):
- Every concept's "Motivation" and "Recap" sub-slides.
- Concept 2's "Common mistake" (prose only — about module path vs directory name).
- The "Practice", "What we learned", and "Up next" sections.

- [ ] **Step 1: Apply the rule mechanically to every entry in the table above**

For each row:
1. Read the existing sub-slide.
2. Move the bulk of the prose explanation onto the parent slide. Where possible, shorten it slightly — the goal of the parent is to be the lead-in, not the whole story.
3. Insert a `--` separator on its own line (with blank lines around it).
4. Start the child slide with `### Code` (or `### Code (broken)` for "Common mistake" sub-slides). Put all code blocks, output blocks, and any post-code explanatory sentences on the child.

Do this for all 11 sub-slides. Don't touch the sub-slides outside the table.

- [ ] **Step 2: Count `--` separators**

```bash
grep -c "^--$" lessons/01-hello/slides/slides.md
```

Expected: `11`. If you see `10` or fewer, you missed one of the rows in the table. If you see `12` or more, a "wrong" sub-slide was split — check Concept 2's "Common mistake" (prose-only) wasn't accidentally split.

- [ ] **Step 3: Verify `---` separators are still present**

The horizontal `---` separators between sub-slides should be untouched.

```bash
grep -c "^---$" lessons/01-hello/slides/slides.md
```

Expected: 26 (unchanged from before — only `--` was added, not `---` removed).

- [ ] **Step 4: Verify every code block now has a `### Code` heading near it**

```bash
grep -c "^### Code" lessons/01-hello/slides/slides.md
```

Expected: at least `11` (one per child slide). Headings like `### Code (broken)` also match.

- [ ] **Step 5: Smoke-test the deck**

```bash
make slides-dev LESSON=01-hello &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/01-hello/slides/slides.md
curl -sS http://localhost:8000/lessons/01-hello/slides/slides.md | grep -c "^--$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `11`.

- [ ] **Step 6: Commit**

```bash
git add lessons/01-hello/slides/slides.md
git commit -m "style(lesson-01): retrofit slides to vertical text → code stacks"
```

---

## Task 10: Retrofit lesson 02 slides.md

Same rule as Task 9, applied to `lessons/02-variables/slides/slides.md`.

**Files:**
- Modify: `lessons/02-variables/slides/slides.md`

### Where the rule applies (lesson 02)

Twelve code-bearing sub-slides need the split:

| Concept | Sub-slide |
|---|---|
| 1. Variables and constants | The basics (includes the `iota` example — keep both code blocks on the child) |
| 1. Variables and constants | A worked example |
| 1. Variables and constants | Common mistake |
| 2. Basic types and zero values | The basics |
| 2. Basic types and zero values | A worked example |
| 2. Basic types and zero values | Common mistake |
| 3. Operators and arithmetic | The basics |
| 3. Operators and arithmetic | A worked example |
| 3. Operators and arithmetic | Common mistake (includes both broken-code and fix-code blocks — keep both on the child) |
| 4. Type conversions | The basics |
| 4. Type conversions | A worked example |
| 4. Type conversions | Common mistake |

**Not split:**
- Every concept's "Motivation" and "Recap" sub-slides.
- The "Practice", "What we learned", and "Up next" sections.

Concept 1's "The basics" is the special case: it has TWO code blocks (the variable forms example AND the `iota` example). The transformation: parent slide carries the prose intro ("Variables hold values..."), child slide carries BOTH code blocks plus the connecting prose between them. The `iota` block doesn't get its own further split.

Concept 3's "Common mistake" has TWO code blocks (broken `total/count` example and the fixed `float64(total)/float64(count)` example). The transformation: parent carries the prose intro ("Integer division silently drops the fractional part:"), child carries BOTH blocks plus the explanation between them.

- [ ] **Step 1: Apply the rule to every entry in the table**

Same procedure as Task 9. Read the sub-slide, move prose to parent, insert `--`, child gets all code with a `### Code` (or `### Code (broken)`) heading.

- [ ] **Step 2: Count `--` separators**

```bash
grep -c "^--$" lessons/02-variables/slides/slides.md
```

Expected: `12`.

- [ ] **Step 3: Verify `---` separators are still present**

```bash
grep -c "^---$" lessons/02-variables/slides/slides.md
```

Expected: 28 (unchanged — your number should match `(grep -c "^---$" lessons/02-variables/slides/slides.md)` before this task started).

- [ ] **Step 4: Verify `### Code` headings**

```bash
grep -c "^### Code" lessons/02-variables/slides/slides.md
```

Expected: at least `12`.

- [ ] **Step 5: Smoke-test the deck**

```bash
make slides-dev LESSON=02-variables &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8000/lessons/02-variables/slides/slides.md
curl -sS http://localhost:8000/lessons/02-variables/slides/slides.md | grep -c "^--$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: `200`, `12`.

- [ ] **Step 6: Commit**

```bash
git add lessons/02-variables/slides/slides.md
git commit -m "style(lesson-02): retrofit slides to vertical text → code stacks"
```

---

## Task 11: End-to-end verification

Confirm everything works as a system: landing page renders correctly, lesson decks load Dracula syntax, vertical drill-down navigation works, lint and tests are clean.

**Files:** none modified — verification only.

- [ ] **Step 1: Run all repo tests**

```bash
make test
```

Expected: every package passes (excluding exercises).

- [ ] **Step 2: Lint clean**

```bash
golangci-lint run ./...
```

Expected: `0 issues.`

- [ ] **Step 3: Build the static site and inspect the landing page**

```bash
make slides-build
test -f dist/index.html && echo OK
grep -c "Phase 1 · Foundations" dist/index.html       # 1
grep -c "Phase 4 · Production" dist/index.html         # 1
grep -c "class=\"lesson future\"" dist/index.html      # 27
grep -c "Hello, Go" dist/index.html                    # ≥ 1
test -f dist/lessons/01-hello/slides/slides.md && echo OK
test -f dist/lessons/02-variables/slides/slides.md && echo OK
test -f dist/shared/reveal/plugin/highlight/dracula.css && echo OK
rm -rf dist
```

Expected: four `OK`s, four matching grep counts (1, 1, 27, ≥1).

- [ ] **Step 4: Serve lesson 01 deck and verify Dracula loads + drill-down works**

```bash
make slides-dev LESSON=01-hello &
SDPID=$!
sleep 2
curl -sS -o /dev/null -w "/lessons/01-hello/slides/index.html: %{http_code}\n" http://localhost:8000/lessons/01-hello/slides/index.html
curl -sS -o /dev/null -w "/shared/reveal/plugin/highlight/dracula.css: %{http_code}\n" http://localhost:8000/shared/reveal/plugin/highlight/dracula.css
curl -sS -o /dev/null -w "/shared/reveal/theme/go-training.css: %{http_code}\n" http://localhost:8000/shared/reveal/theme/go-training.css
curl -sS http://localhost:8000/lessons/01-hello/slides/slides.md | grep -c "^--$"
kill $SDPID
wait $SDPID 2>/dev/null
```

Expected: three `200` lines, then `11`.

- [ ] **Step 5: Manual visual check (browser, optional but recommended)**

```bash
make slides-dev LESSON=01-hello &
```

Open `http://localhost:8000/` in your browser. Confirm:

- Title slide shows "Hello, Go" in white on the gopher-blue background.
- The "LESSON 01" label above the title is Go-green, uppercase, monospace.
- Pressing Right (→) advances through concepts.
- On Concept 1's "The basics" slide, the bottom-right reveal.js arrow indicator shows there's content below.
- Pressing Down (↓) reveals a code-only slide with Dracula syntax highlighting (pink keywords, green types, yellow strings, cyan params, purple symbols).
- Pressing Right again advances to "A worked example" — vertical stack restarts.

Repeat for `make slides-dev LESSON=02-variables`.

Stop the dev server when done.

- [ ] **Step 6: Final repository sanity check**

```bash
git status
make test
git log --oneline main..HEAD
```

Expected: clean working tree, all tests pass, 11 commits on the branch (spec + 10 implementation commits, since Task 11 has no commit).

This task makes no commit — verification only.

---

## Done definition

After Task 11:

- `shared/reveal/theme/go-training.css` is rewritten with the Go-themed palette, Source Sans 3 + Source Code Pro fonts, slide-pattern CSS classes (`.lesson-label`, `.callout`, `.error-block`), and reveal.js variable overrides.
- `shared/reveal/plugin/highlight/dracula.css` is vendored from highlight.js 11.10.0.
- `tools/build-index/main.go` has the static `allLessons` master list of 29 lessons, `buildPhases` logic, and tests covering it.
- `tools/build-index/index.html.tmpl` renders the phase-grouped lesson card grid.
- `tools/new-lesson/template/slides/{index.html.tmpl,slides.md.tmpl}` use Dracula highlighting and pre-stocked vertical `--` separators.
- `CONTRIBUTING.md` documents the "Code goes down" convention.
- `lessons/01-hello/slides/{index.html,slides.md}` and `lessons/02-variables/slides/{index.html,slides.md}` use Dracula and have ~11-12 vertical `--` stacks each.
- `make test` passes; `golangci-lint run ./...` reports 0 issues.
- A push to `main` will redeploy and `https://golang.ristkari.dev/` will show the new theme.

## What ships next (after this PR merges)

- Plan F — Lesson 03 (Control flow). The new scaffolder template (Task 7) makes future lessons start with the right vertical-stack structure for free; no further infrastructure changes needed.
