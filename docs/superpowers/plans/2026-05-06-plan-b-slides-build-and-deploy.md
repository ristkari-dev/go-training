# Plan B — Slides Build + Cloud Run Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `tools/build-index/` (a Go program that produces a static `dist/` directory containing every lesson's deck plus a generated landing page), a multi-stage `Dockerfile` that serves `dist/` over nginx, a Cloud Run service definition, an idempotent GCP setup script, and two GitHub Actions workflows (CI on every push/PR; deploy on push to `main`). Result: a push to `main` builds and deploys the slides to https://golang.ristkari.dev.

**Architecture:** A tiny Go tool walks `lessons/*/slides/`, copies them into `dist/`, copies the vendored `shared/reveal/` into `dist/`, and renders `dist/index.html` from an embedded template using titles parsed from each lesson's README. The Dockerfile builds the tool and runs it during image build, then a non-root nginx image serves the static output. Cloud Run runs the image with no minimum instances. WIF authentication keeps GitHub free of long-lived service-account keys. A custom Cloudflare CNAME points `golang.ristkari.dev` at the Cloud Run domain mapping.

**Tech Stack:** Go 1.23, GNU Make, Docker (multi-stage build), nginx (`nginxinc/nginx-unprivileged:alpine`), Google Cloud Run, Google Artifact Registry, Workload Identity Federation, GitHub Actions (`google-github-actions/auth@v2`, `google-github-actions/setup-gcloud@v2`), Cloudflare DNS (DNS-only mode).

---

## File Structure

After this plan completes:

```
go-training/
├── Makefile                          (modify: add slides-build + slides-docker)
├── tools/
│   └── build-index/                  (Tasks 1-3)
│       ├── main.go
│       ├── main_test.go
│       └── index.html.tmpl           (embedded template for the landing page)
├── deploy/                           (Tasks 4-7)
│   ├── Dockerfile                    multi-stage build → nginx static
│   ├── nginx.conf.template           env-substituted at container start
│   ├── cloudrun.yaml                 Knative-style service spec
│   ├── setup.sh                      one-shot, idempotent GCP bootstrap
│   └── README.md                     one-time setup walkthrough + DNS
├── .github/
│   └── workflows/                    (Tasks 8-9)
│       ├── ci.yml                    push/PR: vet, lint, test, slides-build
│       └── deploy.yml                push to main: build, push image, deploy
└── ...                               (everything from Plan A)
```

### Decomposition rationale

- `tools/build-index/` is split internally: `collectLessons` discovers + sorts, `extractTitle` parses, `copyTree` copies, `renderIndex` produces HTML, `build` orchestrates, `runWithArgs` is the CLI entry. Each piece is independently testable.
- `nginx.conf.template` lives next to the `Dockerfile` so deploy concerns stay in one folder.
- `setup.sh` is a single bash script (per design choice). Idempotency lets it be re-run safely if WIF or AR config drifts.
- The two workflows are split so `ci.yml` runs on PRs without needing GCP secrets, and `deploy.yml` only runs on `main`.

---

## Conventions used by this plan

- **Working directory:** `/Users/ristkari/code/private/go-training/` for every command. Do NOT navigate above this.
- **GCP project ID:** `ristkari-dev`
- **GCP region:** `europe-north1`
- **Cloud Run service name:** `go-training-slides`
- **Artifact Registry repo:** `go-training` (will hold image `slides`)
- **Custom domain:** `golang.ristkari.dev` (DNS hosted at Cloudflare)
- **Service account email:** `github-deploy@ristkari-dev.iam.gserviceaccount.com`
- **WIF pool ID:** `github-actions`
- **WIF provider ID:** `github`
- **GitHub org/repo:** `ristkari-dev/go-training`
- **Commit messages:** Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `test:`, `build:`, `ci:`).
- **Branch:** create a feature branch (e.g. `feature/plan-b-deploy`) for the whole plan; merge to `main` via PR.

---

## Task 1: Bootstrap `tools/build-index/` package + landing-page template

Set up the package skeleton and the embedded HTML template that later tasks render.

**Files:**
- Create: `tools/build-index/main.go` (skeleton)
- Create: `tools/build-index/main_test.go` (skeleton)
- Create: `tools/build-index/index.html.tmpl`

- [ ] **Step 1: Write `tools/build-index/index.html.tmpl`**

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Go Training</title>
  <style>
    :root {
      --bg: #1f2430;
      --fg: #f4f4f4;
      --accent: #5dc9e2;
      --muted: rgba(255, 255, 255, 0.55);
      --code-font: "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
      --main-font: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    }
    body { margin: 0; padding: 3rem 1.5rem; background: var(--bg); color: var(--fg); font: 16px/1.6 var(--main-font); }
    main { max-width: 720px; margin: 0 auto; }
    h1 { font-size: 2.4rem; color: var(--accent); margin: 0 0 0.4rem; letter-spacing: -0.01em; }
    p.lead { color: var(--muted); margin: 0 0 2.5rem; }
    ol.lessons { list-style: none; padding: 0; margin: 0; }
    ol.lessons li { display: flex; align-items: baseline; padding: 0.7rem 0; border-bottom: 1px solid rgba(255,255,255,0.08); }
    ol.lessons li .num { font-family: var(--code-font); color: var(--muted); width: 3.5rem; flex-shrink: 0; }
    ol.lessons li a { color: var(--fg); text-decoration: none; flex: 1; }
    ol.lessons li a:hover { color: var(--accent); }
    .empty { color: var(--muted); padding: 1rem 0; font-style: italic; }
    footer { color: var(--muted); margin-top: 3rem; font-size: 0.9rem; }
    footer a { color: var(--accent); }
  </style>
</head>
<body>
  <main>
    <h1>Go Training</h1>
    <p class="lead">A Go programming course delivered as code + per-lesson reveal.js slide decks.</p>
    {{if .Lessons}}
    <ol class="lessons">
      {{range .Lessons}}
      <li>
        <span class="num">{{.Number}}</span>
        <a href="lessons/{{.Name}}/slides/">{{.Title}}</a>
      </li>
      {{end}}
    </ol>
    {{else}}
    <p class="empty">No lessons published yet.</p>
    {{end}}
    <footer>
      Source: <a href="https://github.com/ristkari-dev/go-training">github.com/ristkari-dev/go-training</a>
    </footer>
  </main>
</body>
</html>
```

- [ ] **Step 2: Write `tools/build-index/main.go` skeleton**

```go
// Command build-index produces a static slides site under dist/: every lesson's
// slides copied into place, the shared reveal.js assets copied alongside, and a
// generated index.html landing page listing all lessons.
package main

func main() {
	// Wired in Task 3.
}
```

- [ ] **Step 3: Write `tools/build-index/main_test.go` skeleton**

```go
package main

import "testing"

// TODO: tests added in Task 2.
var _ = testing.Short
```

The unused-import workaround keeps the file compilable until Task 2 adds real tests.

- [ ] **Step 4: Verify the package builds**

```bash
cd /Users/ristkari/code/private/go-training
go build ./tools/build-index/...
go test ./tools/build-index/... -count=1
```

Expected: `go build` succeeds. `go test` reports `ok ... [no tests to run]` or similar.

- [ ] **Step 5: Commit**

```bash
git add tools/build-index/
git commit -m "feat(build-index): bootstrap package + landing-page template"
```

---

## Task 2: Implement build-index logic with TDD

Implement five small functions that together produce `dist/`. Each gets its own red/green cycle, then a final commit when all tests pass.

**Files:**
- Modify: `tools/build-index/main.go`
- Modify: `tools/build-index/main_test.go`

- [ ] **Step 1: Replace the test file with the full test suite (failing)**

Replace the entire contents of `tools/build-index/main_test.go` with:

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

func TestExtractTitleFromReadme(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, "# Lesson 03: Control Flow\n\nSome content.\n")
	if got := extractTitle(readme, "control-flow"); got != "Control Flow" {
		t.Errorf("got %q, want %q", got, "Control Flow")
	}
}

func TestExtractTitleFallsBackToSlug(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	writeFile(t, readme, "Some text without an H1.\n")
	if got := extractTitle(readme, "slices-and-maps"); got != "Slices And Maps" {
		t.Errorf("got %q, want %q", got, "Slices And Maps")
	}
}

func TestExtractTitleHandlesMissingFile(t *testing.T) {
	if got := extractTitle("/nonexistent/path/README.md", "hello"); got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestCollectLessonsSortedByNumber(t *testing.T) {
	dir := t.TempDir()
	for _, l := range []struct{ name, body string }{
		{"03-control-flow", "# Lesson 03: Control Flow\n"},
		{"01-hello", "# Lesson 01: Hello\n"},
		{"10-modules", "# Lesson 10: Modules\n"},
	} {
		writeFile(t, filepath.Join(dir, l.name, "README.md"), l.body)
		writeFile(t, filepath.Join(dir, l.name, "slides", "index.html"), "<html></html>")
	}
	got, err := collectLessons(dir)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []string{"01-hello", "03-control-flow", "10-modules"}
	if len(got) != len(wantNames) {
		t.Fatalf("got %d lessons, want %d", len(got), len(wantNames))
	}
	for i, w := range wantNames {
		if got[i].Name != w {
			t.Errorf("lesson %d: got %q, want %q", i, got[i].Name, w)
		}
	}
	if got[0].Title != "Hello" {
		t.Errorf("lesson 0 title: got %q, want %q", got[0].Title, "Hello")
	}
}

func TestCollectLessonsIgnoresEntriesWithoutSlides(t *testing.T) {
	dir := t.TempDir()
	// A lesson dir with no slides/ should be skipped.
	writeFile(t, filepath.Join(dir, "01-hello", "README.md"), "# Lesson 01: Hello\n")
	// A lesson dir with slides/.
	writeFile(t, filepath.Join(dir, "02-types", "README.md"), "# Lesson 02: Types\n")
	writeFile(t, filepath.Join(dir, "02-types", "slides", "index.html"), "<html></html>")
	got, err := collectLessons(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "02-types" {
		t.Errorf("got %+v, want only 02-types", got)
	}
}

func TestCollectLessonsHandlesMissingDir(t *testing.T) {
	got, err := collectLessons(filepath.Join(t.TempDir(), "lessons-doesnt-exist"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d lessons, want 0", len(got))
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
	for _, rel := range []string{"a/b/c.txt", "a/d.txt"} {
		body, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			t.Errorf("%s: %v", rel, err)
		}
		if rel == "a/b/c.txt" && string(body) != "hello" {
			t.Errorf("%s body mismatch: %q", rel, body)
		}
	}
}

func TestRenderIndexLists(t *testing.T) {
	body, err := renderIndex([]lesson{
		{Number: "01", Slug: "hello", Name: "01-hello", Title: "Hello"},
		{Number: "02", Slug: "types", Name: "02-types", Title: "Types"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, want := range []string{"Hello", "Types", "lessons/01-hello/slides/", "lessons/02-types/slides/"} {
		if !strings.Contains(s, want) {
			t.Errorf("index missing %q\n---\n%s", want, s)
		}
	}
}

func TestRenderIndexEmpty(t *testing.T) {
	body, err := renderIndex(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "No lessons published yet.") {
		t.Errorf("empty index missing fallback text\n---\n%s", body)
	}
}

func TestBuildEndToEnd(t *testing.T) {
	root := t.TempDir()
	lessonsDir := filepath.Join(root, "lessons")
	sharedDir := filepath.Join(root, "shared", "reveal")
	outDir := filepath.Join(root, "dist")

	writeFile(t, filepath.Join(lessonsDir, "01-hello", "README.md"), "# Lesson 01: Hello\n")
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
	if !strings.Contains(string(idx), "Hello") {
		t.Errorf("index.html missing 'Hello'\n%s", idx)
	}
}
```

- [ ] **Step 2: Run the tests, confirm they fail**

```bash
go test ./tools/build-index/ -v
```

Expected: build failure — `extractTitle`, `collectLessons`, `copyTree`, `renderIndex`, `build`, `lesson` are all undefined.

- [ ] **Step 3: Replace `main.go` with the full implementation**

```go
// Command build-index produces a static slides site under dist/: every lesson's
// slides copied into place, the shared reveal.js assets copied alongside, and a
// generated index.html landing page listing all lessons.
package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed index.html.tmpl
var indexFS embed.FS

type lesson struct {
	Number string
	Slug   string
	Name   string
	Title  string
}

var lessonNameRe = regexp.MustCompile(`^(\d{2})-([a-z][a-z0-9]*(?:-[a-z0-9]+)*)$`)
var titleHeadingRe = regexp.MustCompile(`(?i)^#\s+lesson\s+\d+\s*[:\-]\s*(.+?)\s*$`)

func collectLessons(lessonsDir string) ([]lesson, error) {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []lesson
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := lessonNameRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		slidesDir := filepath.Join(lessonsDir, e.Name(), "slides")
		if info, err := os.Stat(slidesDir); err != nil || !info.IsDir() {
			continue
		}
		readme := filepath.Join(lessonsDir, e.Name(), "README.md")
		out = append(out, lesson{
			Number: m[1],
			Slug:   m[2],
			Name:   e.Name(),
			Title:  extractTitle(readme, m[2]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func extractTitle(readmePath, slug string) string {
	f, err := os.Open(readmePath)
	if err != nil {
		return slugToTitle(slug)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := titleHeadingRe.FindStringSubmatch(scanner.Text()); m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return slugToTitle(slug)
}

func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
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
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func renderIndex(lessons []lesson) ([]byte, error) {
	tmpl, err := template.ParseFS(indexFS, "index.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Lessons []lesson }{lessons}); err != nil {
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
	lessons, err := collectLessons(lessonsDir)
	if err != nil {
		return fmt.Errorf("collect lessons: %w", err)
	}
	for _, l := range lessons {
		src := filepath.Join(lessonsDir, l.Name, "slides")
		dst := filepath.Join(outDir, "lessons", l.Name, "slides")
		if err := copyTree(src, dst); err != nil {
			return fmt.Errorf("copy %s slides: %w", l.Name, err)
		}
	}
	if _, err := os.Stat(sharedDir); err == nil {
		if err := copyTree(sharedDir, filepath.Join(outDir, "shared", "reveal")); err != nil {
			return fmt.Errorf("copy shared/reveal: %w", err)
		}
	}
	idx, err := renderIndex(lessons)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "index.html"), idx, 0o644)
}

func main() {
	// Wired in Task 3.
}
```

- [ ] **Step 4: Run the tests, confirm they pass**

```bash
go test ./tools/build-index/ -v
```

Expected: every test PASSes. `TestRenderIndexLists` may print an `_ = want` line — that's intentional (the loop with `_ = want` is a no-op kept around to make it obvious that the meaningful assertions follow below it).

If a test fails, fix the code (not the test) — the tests encode the spec.

- [ ] **Step 5: Commit**

```bash
git add tools/build-index/main.go tools/build-index/main_test.go
git commit -m "feat(build-index): collect, copy, and render dist site"
```

---

## Task 3: Wire the build-index CLI

**Files:**
- Modify: `tools/build-index/main.go`

- [ ] **Step 1: Replace `func main()` with the full CLI wiring**

Replace the `func main()` stub at the bottom of `tools/build-index/main.go` with:

```go
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

Add `"flag"` to the existing `import (…)` block.

- [ ] **Step 2: Run tests, confirm they still pass**

```bash
go test ./tools/build-index/ -v
```

Expected: every test still PASSes.

- [ ] **Step 3: Smoke-test the binary**

```bash
TMP=$(mktemp -d)
mkdir -p "$TMP/lessons/01-hello/slides" "$TMP/shared/reveal/dist"
echo "# Lesson 01: Hello" > "$TMP/lessons/01-hello/README.md"
echo "<html>hi</html>" > "$TMP/lessons/01-hello/slides/index.html"
echo "console.log('reveal');" > "$TMP/shared/reveal/dist/reveal.js"
go run ./tools/build-index -lessons "$TMP/lessons" -shared "$TMP/shared/reveal" -out "$TMP/dist"
find "$TMP/dist" -type f | sort
rm -rf "$TMP"
```

Expected output ends with four files:

```
<TMP>/dist/index.html
<TMP>/dist/lessons/01-hello/slides/index.html
<TMP>/dist/shared/reveal/dist/reveal.js
```

(Order may vary; the count and paths are what matters.)

- [ ] **Step 4: Commit**

```bash
git add tools/build-index/main.go
git commit -m "feat(build-index): wire CLI with -lessons, -shared, -out flags"
```

---

## Task 4: Add `slides-build` and `slides-docker` targets to the Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Append the two targets to `Makefile`**

After the existing `slides-dev` target, append (preserve TABs!):

```make

.PHONY: slides-build
slides-build: ## Build the static slides site into dist/
	go run ./tools/build-index -lessons lessons -shared shared/reveal -out dist

.PHONY: slides-docker
slides-docker: ## Build the deploy image and run it locally on http://localhost:8080
	docker build -t go-training-slides:local -f deploy/Dockerfile .
	@echo "starting container on http://localhost:8080  (Ctrl-C to stop)"
	docker run --rm -p 8080:8080 -e PORT=8080 go-training-slides:local
```

- [ ] **Step 2: Verify `make help` lists the new targets**

```bash
make help
```

Expected: the listing now includes `slides-build` and `slides-docker` with their descriptions.

- [ ] **Step 3: Verify `make slides-build` works against an empty repo**

```bash
make slides-build
ls dist/
```

Expected: prints `built dist`. `ls dist/` shows only `index.html` (since `lessons/` is empty apart from `.gitkeep`). `dist/index.html` contains "No lessons published yet.".

- [ ] **Step 4: Clean up the dist dir (it's gitignored but tidy)**

```bash
rm -rf dist
```

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "build: add slides-build and slides-docker Makefile targets"
```

---

## Task 5: Write the Dockerfile + nginx config

**Files:**
- Create: `deploy/Dockerfile`
- Create: `deploy/nginx.conf.template`

- [ ] **Step 1: Write `deploy/nginx.conf.template`**

The nginx-unprivileged image's entrypoint substitutes `${VAR}` references in `/etc/nginx/templates/*.conf.template` files using `envsubst`, producing `/etc/nginx/conf.d/*.conf` at startup. Cloud Run injects `PORT=8080` by default.

```nginx
server {
    listen ${PORT} default_server;
    listen [::]:${PORT} default_server;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # Static assets: cache for an hour
    location ~* \.(?:js|mjs|css|svg|woff2?|ttf|otf|png|jpg|jpeg|gif|ico|map)$ {
        expires 1h;
        add_header Cache-Control "public, max-age=3600";
        try_files $uri =404;
    }

    # HTML and markdown: no cache (so deploys propagate immediately)
    location ~* \.(?:html|md)$ {
        add_header Cache-Control "no-cache";
        try_files $uri =404;
    }

    # Default: try the file, then a directory's index.html, else 404
    location / {
        try_files $uri $uri/ =404;
    }

    # Favicon shouldn't 404 noisily if missing
    location = /favicon.ico {
        log_not_found off;
        access_log off;
    }
}
```

- [ ] **Step 2: Write `deploy/Dockerfile`**

```dockerfile
# syntax=docker/dockerfile:1.7

# --- Stage 1: build the static dist/ ---
FROM golang:1.23-alpine AS builder
WORKDIR /src

# Cache deps separately
COPY go.mod ./
RUN go mod download

# Copy everything else and build
COPY . .
RUN go build -o /out/build-index ./tools/build-index \
    && /out/build-index -lessons /src/lessons -shared /src/shared/reveal -out /dist

# --- Stage 2: serve dist/ with non-root nginx ---
FROM nginxinc/nginx-unprivileged:alpine
USER root
COPY deploy/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /dist /usr/share/nginx/html
# Drop back to the unprivileged user the image ships with (UID 101).
USER 101
EXPOSE 8080
```

- [ ] **Step 3: Build the image locally**

```bash
docker build -t go-training-slides:local -f deploy/Dockerfile .
```

Expected: builds successfully. The intermediate `RUN go build && build-index` step prints `built /dist`.

- [ ] **Step 4: Run the image locally and curl it**

```bash
docker run --rm -d -p 8080:8080 -e PORT=8080 --name gts-test go-training-slides:local
sleep 2
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8080/
curl -sS http://localhost:8080/ | head -20
docker stop gts-test
```

Expected:
- First curl: `200`.
- Second curl: HTML containing `<title>Go Training</title>` and the "No lessons published yet." text (since the repo currently has no lessons).

- [ ] **Step 5: Commit**

```bash
git add deploy/Dockerfile deploy/nginx.conf.template
git commit -m "build: add multi-stage Dockerfile + nginx config"
```

---

## Task 6: Write the Cloud Run service definition

**Files:**
- Create: `deploy/cloudrun.yaml`

- [ ] **Step 1: Write `deploy/cloudrun.yaml`**

This is a Knative-style spec. The image tag is replaced at deploy time by the GitHub Actions workflow (Task 9) — a `__IMAGE__` placeholder is substituted with the actual `:sha` tag.

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: go-training-slides
  labels:
    cloud.googleapis.com/location: europe-north1
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "0"
        autoscaling.knative.dev/maxScale: "2"
        run.googleapis.com/cpu-throttling: "true"
    spec:
      containerConcurrency: 80
      timeoutSeconds: 30
      containers:
        - image: __IMAGE__
          ports:
            - name: http1
              containerPort: 8080
          env:
            - name: PORT
              value: "8080"
          resources:
            limits:
              cpu: "1"
              memory: 256Mi
  traffic:
    - percent: 100
      latestRevision: true
```

- [ ] **Step 2: Commit**

```bash
git add deploy/cloudrun.yaml
git commit -m "build: add Cloud Run service definition"
```

---

## Task 7: Write the GCP setup script

A single idempotent bash script that bootstraps everything in the `ristkari-dev` project: enables APIs, creates an Artifact Registry repo, creates the deploy service account, creates a Workload Identity Federation pool + GitHub provider, binds the SA so the GitHub repo can impersonate it, and prints the values to copy into GitHub repo secrets.

**Files:**
- Create: `deploy/setup.sh`

- [ ] **Step 1: Write `deploy/setup.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

# Idempotent GCP bootstrap for the go-training slides deployment.
# Re-runnable: each step either creates the resource or no-ops if it already exists.

PROJECT_ID="${PROJECT_ID:-ristkari-dev}"
REGION="${REGION:-europe-north1}"
AR_REPO="${AR_REPO:-go-training}"
SERVICE_NAME="${SERVICE_NAME:-go-training-slides}"
SA_NAME="${SA_NAME:-github-deploy}"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
WIF_POOL="${WIF_POOL:-github-actions}"
WIF_PROVIDER="${WIF_PROVIDER:-github}"
GITHUB_REPO="${GITHUB_REPO:-ristkari-dev/go-training}"

bold()  { printf '\033[1m%s\033[0m\n' "$*"; }
note()  { printf '  → %s\n' "$*"; }

bold "Project:        $PROJECT_ID"
bold "Region:         $REGION"
bold "AR repo:        $AR_REPO"
bold "Service:        $SERVICE_NAME"
bold "Service acct:   $SA_EMAIL"
bold "WIF pool:       $WIF_POOL"
bold "WIF provider:   $WIF_PROVIDER"
bold "GitHub repo:    $GITHUB_REPO"
echo

bold "1. Enabling required APIs"
gcloud services enable \
    artifactregistry.googleapis.com \
    iamcredentials.googleapis.com \
    run.googleapis.com \
    sts.googleapis.com \
    --project="$PROJECT_ID"

bold "2. Creating Artifact Registry repo (if missing)"
if gcloud artifacts repositories describe "$AR_REPO" \
        --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "repo $AR_REPO already exists, skipping"
else
    gcloud artifacts repositories create "$AR_REPO" \
        --repository-format=docker \
        --location="$REGION" \
        --description="go-training container images" \
        --project="$PROJECT_ID"
fi

bold "3. Creating service account (if missing)"
if gcloud iam service-accounts describe "$SA_EMAIL" --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "service account $SA_EMAIL already exists, skipping"
else
    gcloud iam service-accounts create "$SA_NAME" \
        --display-name="GitHub Actions deploy for go-training" \
        --project="$PROJECT_ID"
fi

bold "4. Granting roles to the service account"
for role in \
    roles/artifactregistry.writer \
    roles/run.admin \
    roles/iam.serviceAccountUser \
; do
    note "binding $role"
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$SA_EMAIL" \
        --role="$role" \
        --condition=None \
        --quiet >/dev/null
done

bold "5. Creating Workload Identity Federation pool (if missing)"
if gcloud iam workload-identity-pools describe "$WIF_POOL" \
        --location=global --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "pool $WIF_POOL already exists, skipping"
else
    gcloud iam workload-identity-pools create "$WIF_POOL" \
        --location=global \
        --display-name="GitHub Actions" \
        --project="$PROJECT_ID"
fi

POOL_NAME=$(gcloud iam workload-identity-pools describe "$WIF_POOL" \
    --location=global --project="$PROJECT_ID" --format='value(name)')

bold "6. Creating WIF OIDC provider for GitHub (if missing)"
if gcloud iam workload-identity-pools providers describe "$WIF_PROVIDER" \
        --location=global --workload-identity-pool="$WIF_POOL" \
        --project="$PROJECT_ID" >/dev/null 2>&1; then
    note "provider $WIF_PROVIDER already exists, skipping"
else
    gcloud iam workload-identity-pools providers create-oidc "$WIF_PROVIDER" \
        --location=global \
        --workload-identity-pool="$WIF_POOL" \
        --display-name="GitHub OIDC" \
        --issuer-uri="https://token.actions.githubusercontent.com" \
        --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref" \
        --attribute-condition="assertion.repository == '${GITHUB_REPO}'" \
        --project="$PROJECT_ID"
fi

PROVIDER_NAME=$(gcloud iam workload-identity-pools providers describe "$WIF_PROVIDER" \
    --location=global --workload-identity-pool="$WIF_POOL" \
    --project="$PROJECT_ID" --format='value(name)')

bold "7. Allowing the GitHub repo to impersonate the SA"
gcloud iam service-accounts add-iam-policy-binding "$SA_EMAIL" \
    --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/${POOL_NAME}/attribute.repository/${GITHUB_REPO}" \
    --project="$PROJECT_ID" \
    --condition=None \
    --quiet >/dev/null

echo
bold "Done. Add these as GitHub repository secrets:"
echo
echo "  GCP_PROJECT_ID                = $PROJECT_ID"
echo "  GCP_WORKLOAD_IDENTITY_PROVIDER = $PROVIDER_NAME"
echo "  GCP_SERVICE_ACCOUNT_EMAIL     = $SA_EMAIL"
echo
bold "Then create the Cloud Run service for the first time (the deploy workflow"
bold "expects the service to exist). One of:"
echo
echo "  # Option A (recommended): apply the checked-in spec with a placeholder image:"
echo "  gcloud run services replace deploy/cloudrun.yaml \\"
echo "      --region=$REGION --project=$PROJECT_ID"
echo "  # (You'll then need to push to main once to deploy a real image — until then"
echo "  #  the service has 'invalid image' status, which is fine.)"
echo
echo "  # Option B: deploy a placeholder image now to materialise the service:"
echo "  gcloud run deploy $SERVICE_NAME \\"
echo "      --image=gcr.io/cloudrun/hello \\"
echo "      --region=$REGION --project=$PROJECT_ID \\"
echo "      --allow-unauthenticated"
echo
bold "Then map the custom domain (Cloudflare DNS will need a CNAME — see deploy/README.md):"
echo "  gcloud beta run domain-mappings create \\"
echo "      --service=$SERVICE_NAME \\"
echo "      --domain=golang.ristkari.dev \\"
echo "      --region=$REGION --project=$PROJECT_ID"
```

- [ ] **Step 2: Make it executable**

```bash
chmod +x deploy/setup.sh
```

- [ ] **Step 3: Sanity-check the script's syntax**

```bash
bash -n deploy/setup.sh && echo OK
```

Expected: `OK`. (Syntax check only — does NOT execute.)

- [ ] **Step 4: Commit**

```bash
git add deploy/setup.sh
git commit -m "build: add idempotent GCP setup script"
```

---

## Task 8: Write `deploy/README.md`

**Files:**
- Create: `deploy/README.md`

- [ ] **Step 1: Write `deploy/README.md`**

```markdown
# Deploying go-training slides

The slides site is built into a Docker image, pushed to Google Artifact
Registry, and served by Cloud Run. A push to `main` triggers
`.github/workflows/deploy.yml` which does the build → push → deploy.

The custom domain `https://golang.ristkari.dev/` points at the Cloud Run
service via a Cloudflare CNAME (DNS-only).

This file documents the **one-time setup** you do once per project (or after
an `iam` cleanup), not what runs on every push.

## Prerequisites

- `gcloud` CLI authenticated as an account with Owner or sufficient roles
  on the `ristkari-dev` project.
- `gh` CLI authenticated against `ristkari-dev/go-training` for setting
  repo secrets (or you can set them in the GitHub UI).
- Cloudflare access for the `ristkari.dev` zone.

## Step 1 — bootstrap GCP

```bash
./deploy/setup.sh
```

The script is idempotent — re-runnable. It:

1. Enables the required APIs (`artifactregistry`, `run`, `iamcredentials`, `sts`).
2. Creates Artifact Registry repo `go-training` in `europe-north1`.
3. Creates service account `github-deploy@ristkari-dev.iam.gserviceaccount.com`.
4. Grants it `roles/artifactregistry.writer`, `roles/run.admin`,
   `roles/iam.serviceAccountUser`.
5. Creates Workload Identity Federation pool `github-actions` and OIDC
   provider `github`, scoped to this repo only.
6. Binds the SA so GitHub Actions on this repo can impersonate it via WIF.
7. Prints the three values you need as GitHub repo secrets.

## Step 2 — set GitHub repo secrets

Take the three values printed by `setup.sh` and set them as repo secrets:

```bash
gh secret set GCP_PROJECT_ID                 -b "ristkari-dev"
gh secret set GCP_WORKLOAD_IDENTITY_PROVIDER -b "<value-from-setup-output>"
gh secret set GCP_SERVICE_ACCOUNT_EMAIL      -b "github-deploy@ristkari-dev.iam.gserviceaccount.com"
```

Or set them in the GitHub UI: **Settings → Secrets and variables → Actions**.

## Step 3 — first-time service materialisation

The deploy workflow uses `gcloud run services replace deploy/cloudrun.yaml`.
That works only if a service named `go-training-slides` already exists. Create
it once with a placeholder image:

```bash
gcloud run deploy go-training-slides \
    --image=gcr.io/cloudrun/hello \
    --region=europe-north1 \
    --project=ristkari-dev \
    --platform=managed \
    --allow-unauthenticated \
    --port=8080
```

After this, the GitHub Actions deploy will replace it with the real image on
every push to `main`.

## Step 4 — map the custom domain

```bash
gcloud beta run domain-mappings create \
    --service=go-training-slides \
    --domain=golang.ristkari.dev \
    --region=europe-north1 \
    --project=ristkari-dev
```

Output includes a CNAME target like `ghs.googlehosted.com.`.

## Step 5 — Cloudflare DNS

Add a CNAME record in the Cloudflare dashboard for `ristkari.dev`:

| Type  | Name   | Target                      | Proxy status |
|-------|--------|-----------------------------|--------------|
| CNAME | golang | `ghs.googlehosted.com`      | **DNS only** (gray cloud) |

Cloudflare proxying (orange cloud) breaks Cloud Run's domain mapping because
Cloud Run handles its own TLS at the mapped hostname. Leave it gray.

Propagation typically takes <5 minutes. Verify with:

```bash
dig golang.ristkari.dev CNAME +short
gcloud beta run domain-mappings describe \
    --domain=golang.ristkari.dev \
    --region=europe-north1 \
    --project=ristkari-dev
```

The `domain-mappings describe` output shows `READY=True` and serves on HTTPS
once Google has provisioned a managed certificate (a few minutes after the
DNS propagates).

## Verifying a deploy

After a push to `main`:

```bash
# Watch the deploy workflow
gh run watch

# When green, hit the service:
curl -sS -I https://golang.ristkari.dev/
```

Expected: `HTTP/2 200`, `content-type: text/html`.

The Cloud Run service URL (without the custom domain) is:

```bash
gcloud run services describe go-training-slides \
    --region=europe-north1 --project=ristkari-dev \
    --format='value(status.url)'
```

## Rolling back

```bash
gcloud run services list --project=ristkari-dev
gcloud run revisions list --service=go-training-slides \
    --region=europe-north1 --project=ristkari-dev
gcloud run services update-traffic go-training-slides \
    --to-revisions=go-training-slides-<previous-revision>=100 \
    --region=europe-north1 --project=ristkari-dev
```
```

- [ ] **Step 2: Commit**

```bash
git add deploy/README.md
git commit -m "docs: add deploy/README with one-time GCP and DNS setup"
```

---

## Task 9: Write the CI workflow

Runs on every push and PR. No GCP secrets needed — this is local-only verification.

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: go vet
        run: go vet ./...

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v2.12.1

      - name: go test (excluding exercises)
        run: make test

      - name: build static slides
        run: make slides-build

      - name: verify dist contents
        run: |
          test -f dist/index.html
          grep -q "Go Training" dist/index.html
```

> The linter version `v2.12.1` matches what was installed locally during Plan A, so CI and local runs report the same findings.

- [ ] **Step 2: Validate the YAML locally**

```bash
python3 -c "import yaml, sys; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo OK
```

Expected: `OK`. (Bash workaround: if `python3` isn't available, `yamllint .github/workflows/ci.yml` works too. The point is to catch indentation errors before pushing.)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add CI workflow (vet, lint, test, slides-build)"
```

---

## Task 10: Write the deploy workflow

Runs on every push to `main`. Uses Workload Identity Federation — no long-lived secrets in GitHub.

**Files:**
- Create: `.github/workflows/deploy.yml`

- [ ] **Step 1: Write `.github/workflows/deploy.yml`**

```yaml
name: Deploy

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  id-token: write

env:
  REGION: europe-north1
  AR_REPO: go-training
  IMAGE_NAME: slides
  SERVICE_NAME: go-training-slides

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - id: auth
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.GCP_WORKLOAD_IDENTITY_PROVIDER }}
          service_account: ${{ secrets.GCP_SERVICE_ACCOUNT_EMAIL }}

      - uses: google-github-actions/setup-gcloud@v2

      - name: Configure Docker for Artifact Registry
        run: gcloud auth configure-docker ${{ env.REGION }}-docker.pkg.dev --quiet

      - name: Build and push image
        id: build
        run: |
          IMAGE="${{ env.REGION }}-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/${{ env.AR_REPO }}/${{ env.IMAGE_NAME }}:${{ github.sha }}"
          docker build -t "$IMAGE" -f deploy/Dockerfile .
          docker push "$IMAGE"
          echo "image=$IMAGE" >> "$GITHUB_OUTPUT"

      - name: Render Cloud Run spec
        run: |
          sed "s|__IMAGE__|${{ steps.build.outputs.image }}|" deploy/cloudrun.yaml > /tmp/cloudrun.rendered.yaml

      - name: Deploy to Cloud Run
        run: |
          gcloud run services replace /tmp/cloudrun.rendered.yaml \
              --region=${{ env.REGION }} \
              --project=${{ secrets.GCP_PROJECT_ID }}

      - name: Allow public access (idempotent)
        run: |
          gcloud run services add-iam-policy-binding ${{ env.SERVICE_NAME }} \
              --region=${{ env.REGION }} \
              --project=${{ secrets.GCP_PROJECT_ID }} \
              --member=allUsers \
              --role=roles/run.invoker || true

      - name: Print service URL
        run: |
          gcloud run services describe ${{ env.SERVICE_NAME }} \
              --region=${{ env.REGION }} \
              --project=${{ secrets.GCP_PROJECT_ID }} \
              --format='value(status.url)'
```

- [ ] **Step 2: Validate the YAML locally**

```bash
python3 -c "import yaml, sys; yaml.safe_load(open('.github/workflows/deploy.yml'))" && echo OK
```

Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: add Cloud Run deploy workflow with WIF auth"
```

---

## Task 11: Local end-to-end smoke test

Verifies the build/serve/CI parts work locally. Does NOT touch GCP — that's the user's manual one-time setup post-merge.

**Files:** none modified.

- [ ] **Step 1: Build a sandbox lesson and `slides-build` against it**

```bash
make new-lesson NAME=99-demo
make slides-build
```

Expected:
- Lesson scaffolded.
- `dist/` created with `dist/index.html`, `dist/lessons/99-demo/slides/*`, `dist/shared/reveal/*`.

```bash
test -f dist/index.html && echo OK
test -f dist/lessons/99-demo/slides/index.html && echo OK
test -f dist/shared/reveal/dist/reveal.js && echo OK
grep -q "99" dist/index.html && grep -q "Demo" dist/index.html && echo OK
```

Expected: four `OK` lines.

- [ ] **Step 2: Build the deploy image locally**

```bash
docker build -t go-training-slides:smoke -f deploy/Dockerfile .
```

Expected: image builds. The build-stage logs show `built /dist`.

- [ ] **Step 3: Run the image and curl it**

```bash
docker run --rm -d -p 8080:8080 -e PORT=8080 --name gts-smoke go-training-slides:smoke
sleep 2

curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8080/
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8080/lessons/99-demo/slides/slides.md
curl -sS -o /dev/null -w "%{http_code}\n" http://localhost:8080/shared/reveal/dist/reveal.js
curl -sS http://localhost:8080/ | grep -q "99" && echo "index lists 99-demo"

docker stop gts-smoke
```

Expected: three `200` lines, then `index lists 99-demo`.

- [ ] **Step 4: Clean up**

```bash
rm -rf lessons/99-demo dist
docker rmi go-training-slides:smoke 2>/dev/null || true
test ! -d lessons/99-demo && test ! -d dist && echo OK
```

Expected: `OK`. `git status` should be clean (the demo lesson was never staged; `dist/` is gitignored).

- [ ] **Step 5: Final repo sanity check**

```bash
git status
make test
```

Expected: clean tree, all tests pass.

This task makes no commit — it is verification only.

---

## Done definition

After Task 11, all of these are true:

- `make slides-build` produces a static `dist/` directory.
- `make slides-docker` builds a working deploy image and serves it on `localhost:8080`.
- `deploy/Dockerfile` builds, the resulting image serves the index page and any scaffolded lesson decks.
- `deploy/setup.sh` exists, is `+x`, and is idempotent.
- `deploy/README.md` describes the one-time setup steps end to end.
- `deploy/cloudrun.yaml` defines the Cloud Run service with the right region/scaling.
- `.github/workflows/ci.yml` runs vet/lint/test/build on every push and PR.
- `.github/workflows/deploy.yml` builds, pushes, and deploys on push to `main` using WIF (no service-account JSON keys in the repo).
- The git history is a clean sequence of small, conventional commits.
- The PR's CI run passes (verifiable once the branch is pushed).

## Manual one-time steps the human runs after merge

These are documented in `deploy/README.md` and are NOT part of any task above. The workflow on `main` will fail until the user does them at least once:

1. Run `./deploy/setup.sh`.
2. Set the three GitHub repo secrets (script prints exact values).
3. Materialise the Cloud Run service once with a placeholder image (script prints the gcloud command).
4. Map the custom domain `golang.ristkari.dev` (gcloud beta command).
5. Add the CNAME in Cloudflare (DNS-only, gray cloud).

After that, every push to `main` redeploys automatically, and `https://golang.ristkari.dev/` serves the latest slides.

## Out of scope (deferred to later plans)

- Lesson content (Plans C-F).
- Live-reload for the local slides dev server (Plan A's `make slides-dev` is good enough as-is).
- Preview deploys per PR (could add later as a workflow_dispatch variant).
- Slack/email notifications on deploy success/failure.
- Image vulnerability scanning (Trivy/Grype) in CI — straightforward to add later.
- Cloud Run min-instances > 0 (cost optimisation; we keep min=0 because the audience is small).
