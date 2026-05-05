# Go Training Course — Design

**Status:** Approved (brainstorming complete, awaiting implementation plan)
**Date:** 2026-05-05
**Owner:** Aki Ristkari

## Summary

A Go programming course delivered as both a code repository and per-lesson reveal.js slide decks. The arc starts at "programming 101" for software engineering students with no prior production language experience and finishes with advanced concurrency, systems programming, production services, tooling, and distributed patterns suitable for early-career and experienced developers. Hybrid delivery (live + self-study) over one to two semesters.

## Audience and delivery

- **Audience:** Software engineering students at the start, early-career and experienced engineers by the end. The same arc serves both, with optional "going further" material per lesson for stronger students.
- **Delivery:** Hybrid. Some lessons live (lectures + live coding), some self-study. Each lesson is sized for ~90 minutes of live time plus self-study work.
- **Length:** 1-2 semesters (~24-28 lessons).
- **Language:** English only.

## Hands-on model

- Per-lesson exercises in a starter repo students clone.
- Each exercise ships **starter code that compiles** plus **failing `*_test.go` tests** that act as the spec. Students make the tests pass.
- Solutions ship in the same lesson folder under `solutions/`, committed to `main`. Students can peek if stuck.
- Testing is woven through every lesson from lesson 4 onward.

## Curriculum (28 lessons, four phases)

### Phase 1 — Foundations (lessons 1-7)

1. **Hello, Go** — install, `go run`, `go mod init`, `main`, packages, `fmt`. Print + read input.
2. **Variables, types, operators** — primitives, zero values, type inference, conversions, constants, `iota`.
3. **Control flow** — `if`, `for` (the only loop), `switch` including a type-switch teaser, early returns.
4. **Functions & first tests** — multi-return, named returns, variadic, `defer`. Intro to `*_test.go`. Table tests start here.
5. **Composite types I — arrays, slices, maps** — slice internals (len/cap), `append`, `range`, map idioms.
6. **Composite types II — structs & methods** — value vs pointer receivers, struct embedding basics.
7. **Packages & modules** — splitting code, exported vs unexported, `go.mod`, imports, std-lib tour. Tooling thread starts: `gofmt`, `go vet`.

### Phase 2 — Idiomatic Go (lessons 8-14)

8. **Pointers, value vs reference semantics** — when to use pointers, escape-analysis intuition (no internals).
9. **Interfaces** — implicit satisfaction, small interfaces, `io.Reader`/`io.Writer`, `any`, type assertions.
10. **Errors** — sentinel, wrapping (`errors.Is`/`As`), custom error types, error design.
11. **Generics** — type parameters, constraints, when *not* to use generics.
12. **Encoding & I/O** — JSON, files, readers/writers, `bufio`, streaming patterns.
13. **Time, strings, bytes, regex** — practical std-lib literacy.
14. **Idiomatic project structure & testing patterns** — `cmd/`, `internal/`, table tests, subtests, `t.Helper`, golden files.

### Phase 3 — Concurrency & Systems (lessons 15-21)

15. **Goroutines & channels — the basics** — `go`, unbuffered/buffered channels, `range` over channel, `close`.
16. **Select & timers** — `select`, `time.After`, ticker, default branch, cancellation patterns.
17. **`sync` & memory model** — `Mutex`, `RWMutex`, `WaitGroup`, `Once`, atomics, race detector.
18. **`context`** — propagation, cancellation, deadlines, common mistakes.
19. **Concurrency patterns** — worker pool, fan-in/out, pipelines, `errgroup`.
20. **Networking & syscalls** — `net`, TCP/UDP basics, signals, file descriptors, syscall awareness.
21. **Profiling, benchmarking, fuzzing** — `go test -bench`, pprof, escape analysis, fuzz tests.

### Phase 4 — Production & Distributed (lessons 22-28)

22. **HTTP servers** — `net/http`, routing, handlers, middleware, structured logging (`slog`).
23. **HTTP clients & resilience** — timeouts, retries, circuit-breaker concept, context propagation.
24. **gRPC** — protobuf, server/client, streaming, interceptors.
25. **Configuration, secrets, graceful shutdown** — flags, env, config files, signal handling.
26. **Build, release, containerization** — build tags, cross-compile, multi-stage Dockerfile, distroless.
27. **Observability** — metrics, traces (OpenTelemetry), structured logs, health checks.
28. **Distributed patterns & wrap-up** — message queues, idempotency, deployment to a managed runtime, course capstone.

### Cross-cutting threads

- **Testing** — every lesson from lesson 4 onward ships failing tests for students to make pass.
- **Tooling** — `go vet`, `gofmt`, `goimports`, `staticcheck`, `golangci-lint` introduced lesson 7 and reinforced in CI.
- **"Going further"** — every lesson README has a section with optional advanced exercises, std-lib reading, and external links for stronger students.

## Repository layout

```
go-training/
├── README.md
├── go.mod                       # single module: github.com/<owner>/go-training
├── go.sum
├── Makefile
├── .golangci.yml
├── .editorconfig
├── .gitignore
│
├── lessons/
│   ├── 01-hello/
│   │   ├── README.md            # self-study notes + "going further"
│   │   ├── slides/
│   │   │   ├── index.html
│   │   │   ├── slides.md
│   │   │   └── assets/
│   │   ├── exercises/
│   │   │   ├── hello.go
│   │   │   └── hello_test.go
│   │   └── solutions/
│   │       ├── hello.go
│   │       └── hello_test.go
│   ├── 02-variables/
│   ├── ...
│   └── 28-distributed/
│
├── shared/
│   └── reveal/                  # vendored reveal.js + theme + plugins
│       ├── dist/
│       ├── plugin/
│       └── theme/
│
├── deploy/
│   ├── Dockerfile               # nginx serving all decks + index page
│   ├── nginx.conf
│   ├── cloudrun.yaml
│   └── README.md                # one-time GCP setup
│
├── tools/
│   ├── build-index/             # generates dist/index.html listing every lesson
│   ├── slides-dev/              # local dev server for one deck
│   └── new-lesson/              # scaffolds a new lesson from template
│
├── docs/
│   └── superpowers/specs/       # design docs
│
└── .github/
    └── workflows/
        ├── ci.yml
        └── deploy.yml
```

### Module strategy

- Single `go.mod` at the root. Module path `github.com/<owner>/go-training`.
- Each lesson's `exercises/` and `solutions/` are independent packages under that module.
- No cross-lesson imports — each lesson is self-contained, so renumbering or rewriting a lesson never breaks another.
- `go test ./...` from the root runs every lesson's tests in CI in one invocation.

## Anatomy of a lesson

Every `lessons/NN-name/` folder contains four parts:

### `README.md`

1. **Learning goals** — 3-5 bullets.
2. **Prereqs** — links to earlier lessons.
3. **Concepts** — 1-3 paragraphs of prose mirroring the deck narrative for self-study.
4. **Exercise brief** — what to build, what `go test ./...` should show when done.
5. **How to run** — `cd lessons/NN-name/exercises && go test ./...` (and `go run .` when applicable).
6. **Going further** — optional advanced material.

### `slides/`

- `index.html` — minimal reveal.js bootstrap referencing `../../shared/reveal/`, configures the markdown plugin, points at `slides.md`.
- `slides.md` — markdown content with `---` horizontal separators and `--` vertical separators (used sparingly). Code in fenced `go` blocks. Speaker notes via `Note:` blocks. HTML escape hatches for layout, fragments, and two-column slides.
- `assets/` — diagrams (SVG preferred), images.

### `exercises/`

- One or more `.go` files with **compiling but incomplete** code: function signatures present, bodies return zero values or `panic("TODO")`.
- One or more `*_test.go` files with **failing tests**. Tests are the spec.
- Imports kept minimal; no third-party deps unless the lesson explicitly introduces one.

### `solutions/`

- Same package shape as `exercises/`, fully implemented.
- Tests are identical to those in `exercises/` so swapping in `solutions/` and re-running `go test` passes.
- Brief `// why:` comments only where a choice is non-obvious.

### Conventions

- File names lowercase with underscores. Test files always `_test.go`.
- Each lesson is a single Go package (no sub-packages until late lessons).
- Lesson folder name `NN-kebab-case` with two-digit numbering (`01`, `02`, …, `28`) so listings sort naturally.

## Slide deck workflow

### Authoring

- Markdown with HTML escape hatches.
- Code-heavy slides limit to ~15 visible lines; longer examples split across slides with `[highlight]` annotations to focus attention.
- First slide of every deck: lesson number, title, one-line learning goal.
- Last slide: a "what's next" pointer to the next lesson.
- Diagrams as SVG; never images of code.

### Shared assets

- `shared/reveal/` holds a pinned, vendored reveal.js — no CDN.
- `shared/reveal/theme/go-training.css` is a custom theme tuned for code-heavy decks: monospace at readable size, Go gopher accent colour, generous code-block padding, no distracting transitions.
- Default plugins: `markdown`, `highlight`, `notes`, `search`. Anything else is a per-deck escape hatch.

### Per-deck `index.html`

- Identical scaffolding across decks. Loads `../../shared/reveal/dist/reveal.css`, the theme CSS, and a single `<section data-markdown="slides.md" data-separator="^---$" data-separator-vertical="^--$">`.
- `<head>` sets the deck title from the lesson name.
- Generated by `make new-lesson` so authors don't copy-paste.

### Local development

- `make slides-dev LESSON=NN-name` starts a local static server (`go run ./tools/slides-dev`) on `localhost:8000` serving the requested lesson.
- A server is needed because reveal.js's markdown plugin uses `fetch`, which doesn't work over `file://`.
- Live reload is optional — drop it if it adds complexity.

### Build for deployment

- `make slides-build` runs `tools/build-index/`, which:
  1. Walks `lessons/*/slides/`.
  2. Generates `dist/index.html` listing every lesson with title and link.
  3. Copies each `slides/` directory and the shared `reveal/` assets into `dist/`, preserving the `lessons/NN/slides/` URL shape.
- Output is fully static. No npm, no Node — Go + vendored reveal.js is the entire toolchain.

## Deployment workflow (Docker + Cloud Run)

### Container image

`deploy/Dockerfile` is two stages:

1. **Build stage** — `golang:1.23-alpine` (or current). Runs `make slides-build` to produce the static `dist/`.
2. **Runtime stage** — `nginxinc/nginx-unprivileged:alpine` (Cloud Run requires non-root). Copies `dist/` into `/usr/share/nginx/html/`. Uses `deploy/nginx.conf` which:
   - Listens on `$PORT` (Cloud Run injects this; default `8080`).
   - Serves static files with `Cache-Control: public, max-age=3600` for assets and `no-cache` for HTML.
   - Returns `404` for unknown paths (no SPA fallback).

### Cloud Run service

- One service (`go-training-slides`) in a single region.
- Public ingress, unauthenticated.
- Min instances `0`, max `1-2`. 256 MiB / 1 vCPU.
- `cloudrun.yaml` (Knative-style service definition) checked in for reproducibility.

### CI/CD

Two GitHub Actions workflows under `.github/workflows/`:

**`ci.yml`** (push/PR):
- `go vet ./...`
- `golangci-lint run`
- `go test ./...` — covers both `exercises/` and `solutions/` so we catch broken specs or solution drift.
- `make slides-build` — verifies the static site builds cleanly.

**`deploy.yml`** (push to `main`):
- Builds the Docker image.
- Authenticates to GCP via Workload Identity Federation (no long-lived service account keys).
- Pushes the image to Artifact Registry.
- Deploys to Cloud Run using `cloudrun.yaml`.

### One-time setup (documented in `deploy/README.md`)

- GCP project + Artifact Registry repo + Cloud Run service.
- WIF pool + GitHub provider mapped to this repo.
- GitHub repo secrets: `GCP_PROJECT_ID`, `GCP_WORKLOAD_IDENTITY_PROVIDER`, `GCP_SERVICE_ACCOUNT_EMAIL`. No JSON keys.

### Local parity

- `make slides-docker` builds the same image locally and runs it on `localhost:8080`.

## Tooling and scaffolding

### Makefile

The single canonical entry point. Self-documenting via `make help`:

| Target | What it does |
|---|---|
| `make help` | Lists all targets |
| `make test` | `go test ./...` |
| `make test-lesson LESSON=NN-name` | Tests for one lesson |
| `make lint` | `golangci-lint run` |
| `make fmt` | `gofmt -w` + `goimports -w` |
| `make slides-dev LESSON=NN-name` | Local server for one deck |
| `make slides-build` | Builds the full static `dist/` |
| `make slides-docker` | Builds and runs the deploy image locally on `:8080` |
| `make new-lesson NAME=NN-name` | Scaffolds a new lesson |

### Lesson scaffolder

`tools/new-lesson/` (a Go program):

- Refuses to overwrite an existing folder.
- Copies a `tools/new-lesson/template/` tree (`README.md`, `slides/index.html`, `slides/slides.md`, `exercises/main.go`, `exercises/main_test.go`, `solutions/main.go`, `solutions/main_test.go`).
- Substitutes the lesson name and number where templated.

### Linting

- `.golangci.yml` enables a conservative set: `govet`, `staticcheck`, `errcheck`, `gosimple`, `ineffassign`, `unused`, `gofmt`, `goimports`. No nitpicky linters that fight with teaching examples.
- `.editorconfig` for whitespace consistency.

### Versioning & dependencies

- Go version pinned in `go.mod` and matched in CI + Dockerfile.
- Reveal.js vendored at a pinned version under `shared/reveal/`. Upgrades are explicit, reviewable commits.
- Third-party Go deps avoided in early/middle lessons. Introduced deliberately and only when a lesson teaches them (e.g., `golang.org/x/sync/errgroup` in lesson 19, gRPC packages in lesson 24).

### Documentation in the repo

- Top-level `README.md` — what the course is, prerequisites, install Go + clone + run first lesson.
- `CONTRIBUTING.md` — for adding lessons: the four-file convention, `make new-lesson`, slide style.
- `deploy/README.md` — one-time GCP setup.

## Non-goals

- **No npm / no Node toolchain.** Slides build with Go + vendored static reveal.js.
- **No web framework lessons.** Course teaches `net/http` and stdlib; Gin/Echo/Fiber out of scope.
- **No databases.** Per audience scope decision.
- **No Kubernetes.** Lesson 26 covers containers; deployment story stops at Cloud Run-style managed runtimes.
- **No parallel "experienced developer" deck per lesson.** Stronger students are served by per-lesson "Going further" sections.
- **No auto-grading service.** Tests in `exercises/` are the spec; students self-verify with `go test`.
- **No video recording / streaming infrastructure.** Course is hybrid — that's a delivery concern, not a repo concern.

## Open items deferred to implementation planning

- Exact Go version pin (current stable at implementation time).
- Exact reveal.js version pin (current stable at implementation time).
- Exact GCP region for Cloud Run (depends on owner preference).
- Whether to use Terraform for the GCP setup or a one-shot bash script in `deploy/setup.sh` (default: bash script, swappable later).
- Whether `make slides-dev` includes live reload (default: no, add only if friction warrants).
