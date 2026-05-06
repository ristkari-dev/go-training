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
