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
   Sections: Learning goals, Prerequisites, Concepts, Exercise: warm-up, Exercise: main, How to run,
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

**Code goes "down."** When a sub-slide has both explanatory prose and a code block, split them into a vertical stack with `--`. The prose is the parent slide; the code is the child below it. Authors give each vertical step its own `###` sub-heading (e.g. `### Code`, `### Run it`, `### Output`). Students press Right to move between concepts; Down to drill into code. Reveal.js shows a navigation arrow at the bottom-right when there's more below — no need for an explicit "press Down" cue.

### When students start writing tests

Through lessons 1-3, exercises ship pre-written failing tests that students don't author themselves — they just edit the `.go` files until the tests pass. From **lesson 4 onward**, students write some test code themselves.

### Going further

Each lesson's README ends with a `## Going further` section split into two parts:

- **Read** — 1-3 short links (Go blog posts, std-lib docs, occasional book references).
- **Try** — 1-2 stretch problems harder than the main exercise. Self-graded; no reference solutions in `solutions/`. The point is to push past the spec.

The scaffolder's `README.md.tmpl` pre-stocks both subheadings.

### Imports

Phase 1 stays on the standard library only. The root `go.mod`'s `require` block stays empty through all of Phase 1. Third-party dependencies enter Phase 2 onward, and only when a lesson teaches them.

### Reference

- [Course design](docs/superpowers/specs/2026-05-05-go-course-design.md)
- [Phase 1 design (Plan C spec)](docs/superpowers/specs/2026-05-07-plan-c-phase-1-foundations-design.md)

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
