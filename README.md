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
