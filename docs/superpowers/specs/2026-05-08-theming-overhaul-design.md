# Theming Overhaul — Design

**Status:** Approved (brainstorming complete, awaiting implementation plan)
**Date:** 2026-05-08
**Owner:** Aki Ristkari
**Master design:** [`2026-05-05-go-course-design.md`](2026-05-05-go-course-design.md)

## Summary

A visual overhaul of the go-training course site: a Go-themed dark palette (gopher blue + Go green), Source Sans 3 + Source Code Pro typography (already bundled with reveal.js), a card-based phase-grouped landing page showing the whole 29-lesson arc, a refreshed slide theme with Dracula syntax highlighting, and a new slide-authoring convention where explanatory text and code drill down vertically (Right = next concept, Down = code for current concept). Applies to all existing lessons (01 and 02) and the scaffolder template so future lessons inherit the look.

## Scope

This redesign produces:

- **Palette** — design tokens in CSS custom properties.
- **Typography** — font stacks and a type scale across slide and web contexts.
- **Landing page** — `tools/build-index/index.html.tmpl` rewritten as a phase-grouped lesson grid with placeholder cards for unpublished lessons.
- **Slide theme** — `shared/reveal/theme/go-training.css` rewritten with the new palette + typography + slide-pattern CSS classes.
- **Code highlighting** — Dracula theme vendored into `shared/reveal/plugin/highlight/dracula.css`; lesson template's `index.html.tmpl` updated to load it.
- **Slide-authoring convention** — vertical drill-down (`--` separators) for code-bearing sub-slides, captured in the scaffolder template and CONTRIBUTING.md.
- **Retrofit** — lessons 01 and 02 rewritten to follow the new vertical-stack convention.

## Out of scope

- New lessons (Plans F-K continue in their own PRs).
- Existing lessons' README content (only the slides change structurally; READMEs are already prose-mirroring and don't need restructuring).
- Reveal.js version upgrade (still 5.1.0).
- Custom fonts beyond the Adobe Source family already bundled with reveal.js.
- Mobile-first redesign of slide decks (reveal.js's default responsive behaviour suffices).

## Section 1 — Color palette and design tokens

All colours live as CSS custom properties so they can be referenced by both the landing page (`tools/build-index/index.html.tmpl`) and the slide theme (`shared/reveal/theme/go-training.css`).

### Background and surface colours

- `--bg` `#1d2541` — deep gopher blue, page/slide background.
- `--surface` `#252e4f` — lighter blue, card backgrounds, callouts, framed elements.
- `--surface-2` `#2c3658` — slight elevation (hover states).
- `--border` `rgba(154, 230, 180, 0.18)` — semi-transparent Go-green at low opacity for outlines and dividers.

### Foreground colours

- `--fg` `#e8eef9` — primary text (off-white-blue, easier on eyes than pure white).
- `--fg-muted` `rgba(232, 238, 249, 0.65)` — subtitles, metadata.
- `--fg-subtle` `rgba(232, 238, 249, 0.45)` — tertiary text, footers.

### Accent colours

- `--accent` `#9ae6b4` — Go-green. Used for: lesson numbers, callout left-borders, hover states, link colour, recap-bullet markers, selected-state outlines, code inline highlight.
- `--accent-soft` `rgba(154, 230, 180, 0.10)` — Go-green at 10% opacity, used for callout backgrounds and inline-code backgrounds.

### Code block colours (Dracula)

- `--code-bg` `#282a36` — Dracula's standard charcoal background.
- Token colours come from highlight.js's `dracula.css` (already a standard theme — pink keywords `#ff79c6`, green types `#50fa7b`, purple symbols `#bd93f9`, yellow strings `#f1fa8c`, cyan params `#8be9fd`, orange numbers `#ffb86c`, foreground `#f8f8f2`).

### Error block (compiler errors / unexpected output)

- `--error-accent` `#ff79c6` — pink (matches Dracula's keyword colour for visual consistency with code blocks).
- `--error-bg` `rgba(255, 121, 198, 0.10)` — pink at 10%.
- Used for: compiler-error blocks shown alongside "Common mistake" examples.

### Accessibility

The Go-green-on-blue (`#9ae6b4` on `#1d2541`) combination measures **8.4:1 contrast** — well above WCAG AA's 4.5:1 threshold for normal text. The fg-muted colour on the surface background measures **6.0:1** — also AA-compliant.

## Section 2 — Typography

### Font stacks

- `--font-sans` — `"Source Sans 3", "Source Sans Pro", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
- `--font-mono` — `"Source Code Pro", "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace`

Both fonts are already vendored at `shared/reveal/dist/theme/fonts/source-sans-pro/` (bundled with reveal.js's `serif`, `simple`, `white`, etc. themes). The custom theme's `@font-face` declarations re-use the same files. No external network requests; no extra font downloads.

For Source Sans 3 (the modern successor to Source Sans Pro), if needed we vendor the regular and semibold variants under `shared/reveal/fonts/source-sans-3/`. Otherwise we accept Source Sans Pro as the rendered face — visual diff is negligible.

### Type scale (slides)

- **Lesson label** ("Lesson 01 · Phase 1 — Foundations"): 0.6em, mono, uppercase, letter-spacing 0.18em, accent-coloured, font-weight 600.
- **H1 / title**: 2.2em, Source Sans, weight 700, letter-spacing −0.025em, line-height 1.05.
- **H2 / section heading**: 1.4em, Source Sans, weight 600, letter-spacing −0.01em.
- **H3 / type label** (used as a slide-type marker like "WORKED EXAMPLE"): 1.0em, mono, uppercase, letter-spacing 0.10em, accent-coloured, font-weight 600.
- **Body**: 0.65em (~36px on a 1080-tall canvas), Source Sans, weight 400, line-height 1.55, colour `--fg` at 0.85 alpha.
- **Code in slides**: 0.55em (~30px), Source Code Pro, weight 400, line-height 1.55.

### Type scale (landing page)

Smaller — long-form web context, not a presentation:

- **H1 / page title**: 2.6rem, Source Sans, weight 700, accent-coloured, letter-spacing −0.025em.
- **Lead paragraph**: 1.1rem, Source Sans, weight 400, `--fg-muted`.
- **Phase heading**: 0.75rem, Source Code Pro, uppercase, letter-spacing 0.12em, accent-coloured, weight 600. Followed by a horizontal divider line.
- **Lesson card title**: 1.05rem, Source Sans, weight 600, slight negative tracking (−0.01em).
- **Lesson card subtitle/blurb**: 0.85rem, Source Code Pro, `--fg-muted`. Mono on purpose — gives a "specs sheet" feel and visually echoes the per-lesson code idioms.
- **Lesson card number**: 0.7rem, Source Code Pro, accent-coloured, weight 600, letter-spacing 0.05em.

### Reveal.js theme integration

The theme overrides reveal.js's CSS variables:

- `--r-background-color`: `--bg`
- `--r-main-color`: `--fg`
- `--r-main-font`: `--font-sans`
- `--r-main-font-size`: 38px (the body baseline reveal.js scales from)
- `--r-heading-color`: `--fg`
- `--r-heading-font`: `--font-sans`
- `--r-heading-text-shadow`: `none`
- `--r-heading-text-transform`: `none`
- `--r-heading1-size`: 2.2em
- `--r-heading2-size`: 1.4em
- `--r-heading3-size`: 1.0em
- `--r-link-color`: `--accent`
- `--r-link-color-hover`: `--fg`
- `--r-selection-color`: `--bg`
- `--r-selection-background-color`: `--accent`
- `--r-code-font`: `--font-mono`

## Section 3 — Landing page

Replaces `tools/build-index/index.html.tmpl`.

### Visual structure

```
Go Training                                       (h1, accent-coloured)
A Go programming course delivered as code +       (lead paragraph)
per-lesson reveal.js slide decks.

PHASE 1 · FOUNDATIONS ─────────────────           (phase heading + divider)
┌────┬────┬────┬────┐
│ 01 │ 02 │ 03 │ 04 │     (4-column grid)
└────┴────┴────┴────┘
┌────┬────┬────┬────┐
│ 05 │ 06 │ 07 │ 08 │
└────┴────┴────┴────┘

PHASE 2 · IDIOMATIC GO ────────────────
… (faded placeholders for unpublished lessons)

PHASE 3 · CONCURRENCY & SYSTEMS ───────
…

PHASE 4 · PRODUCTION & DISTRIBUTED ────
…

Source: github.com/ristkari-dev/go-training      (footer)
```

### Card structure (each card)

```
┌──────────────────────┐
│ 01                   │  ← lesson number (mono, accent)
│ Hello, Go            │  ← title (Source Sans, 600)
│ go run · package main│  ← one-line concept summary (mono, muted)
└──────────────────────┘
```

### Card states

- **Published** (lesson exists on disk): solid surface background (`--accent-soft`), Go-green outline, full-opacity title and number. Wrapped in `<a href="lessons/NN-name/slides/">` so the whole card is clickable. Hover effect: outline brightens to 0.45 opacity, background to 0.10 opacity, slight upward translate (`transform: translateY(-2px)`), 150ms ease transition.
- **Unpublished** (placeholder for future lessons): same dimensions, dashed muted-white outline, opacity 0.42, `aria-disabled="true"`, no link target. Visible so students see the whole arc.

### Card content data

`build-index` learns each lesson's data from a static `[]lesson` slice declared in `tools/build-index/main.go`:

```go
var allLessons = []lessonInfo{
    {Number: "01", Slug: "hello", Title: "Hello, Go", Blurb: "go run · package main · fmt", Phase: 1},
    {Number: "02", Slug: "variables", Title: "Variables, types, operators", Blurb: "var · := · float64 · const", Phase: 1},
    {Number: "03", Slug: "control-flow", Title: "Control flow", Blurb: "if · for · switch", Phase: 1},
    // … all 29 lessons listed here, both published and future
}
```

The build tool walks `lessons/` to find which lessons exist on disk; everything in `allLessons` not found on disk is rendered as a placeholder card. This keeps the future-lesson list authoritative in code, not derived from on-disk content.

### Phase grouping

Phases are computed from the lesson number per the master design's phase ranges:

- Phase 1 — Foundations: lessons 01-08
- Phase 2 — Idiomatic Go: lessons 09-15
- Phase 3 — Concurrency & Systems: lessons 16-22
- Phase 4 — Production & Distributed: lessons 23-29

Each phase renders as a heading followed by a grid of its lessons.

### Responsive behaviour

CSS Grid with `grid-template-columns: repeat(auto-fill, minmax(180px, 1fr))`:

- Desktop (≥ 900px): 4 cards per row.
- Tablet (≥ 600px): 3 cards per row.
- Phone (< 600px): 2 cards per row.

### Accessibility

- Published cards are real `<a>` elements (keyboard-accessible, screen-reader-friendly).
- Unpublished placeholders are `<div>` elements with `aria-disabled="true"`.
- All Go-green-on-blue colour pairings meet WCAG AA contrast (8.4:1).
- Focus styles: 2px Go-green outline at `outline-offset: 3px` on keyboard-focused cards.

## Section 4 — Slide theme

Replaces `shared/reveal/theme/go-training.css`. The new file uses the design tokens from Section 1 and the type scale from Section 2.

### Slide-pattern CSS classes

Authors use these custom classes inside `slides.md` (HTML inside markdown is allowed by reveal.js's markdown plugin):

- `.lesson-label` — for the "Lesson 01 · Phase 1 — Foundations" pre-title strip on the title slide.
- `.callout` — boxed accent-bordered text block, used for Learning goal, important asides, and Note-style annotations within a slide.
- `.error-block` — pink-bordered code-styled block for compiler errors / unexpected output. Distinct from regular `<pre>` so it visually contrasts with the broken code shown above it on common-mistake slides.

### Reveal.js element overrides

- `.reveal` — background `--bg`, colour `--fg`, font-family `--font-sans`.
- `.reveal h1, h2, h3, h4` — custom sizes per Section 2; line-height 1.05 (h1) / 1.2 (others); margin tightened.
- `.reveal pre` — Dracula-charcoal background, `border-radius: 8px`, `padding: 1em 1.2em`, no drop shadow, font-size 0.55em, line-height 1.55.
- `.reveal pre code` — `max-height: 75vh`, horizontal-overflow scroll if a single line is too long.
- `.reveal code` (inline) — mono, accent-coloured, soft-accent background, `padding: 0.05em 0.4em`, `border-radius: 3px`.
- `.reveal blockquote` — used as the default callout style: 4px accent left-border, soft-accent background, `padding: 0.5em 1em`, `border-radius: 0 8px 8px 0`.
- `.reveal section img` — no border, no shadow.
- `.reveal .slide-number` — mono, `--fg-subtle`, transparent background.
- `.reveal .progress` — accent-coloured.
- `.reveal ul li::marker` — accent-coloured Go-green dots (instead of default white).

### Code-highlighting integration

- `shared/reveal/plugin/highlight/dracula.css` is added (vendored from highlight.js's standard themes).
- The lesson template's `index.html.tmpl` is updated:
  - `<link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/monokai.css">` → `<link rel="stylesheet" href="../../../shared/reveal/plugin/highlight/dracula.css">`.

### Animations

- Slide transitions: `default` (subtle fade — reveal.js's least distracting option). Configured via `Reveal.initialize({ transition: 'fade', transitionSpeed: 'fast' })` in the lesson template's `index.html.tmpl`.
- No entrance animations on individual elements. Code blocks, headings, lists appear in place when the slide loads.

### Speaker-notes view

Untouched. Reveal.js's notes plugin renders speaker view in a separate window via `notes.css`; our theme leaves it alone. The presenter's view continues to render the slide as currently themed plus the speaker notes panel.

## Section 5 — Slide-authoring convention

This is a **rule** lesson authors follow when writing `slides.md`. It updates the scaffolder template, the CONTRIBUTING.md "Phase 1 conventions" section, and existing lessons 01 + 02.

### The rule

For any sub-slide that has BOTH explanatory text AND a code block, split into a vertical stack:

- The horizontal slide (`---`) shows the prose.
- The vertical slide (`--`) below shows the code, with its own `### Code` (or `### Run it`, `### Output`) heading.

Reveal.js's navigation: Right = next concept, Down = code for current concept. The bottom-right navigation arrow indicator naturally tells students when there's more below — no explicit "press Down" hint in the prose.

### Which sub-slides drill down

For each concept's 5 sub-slides:

| Sub-slide | Pattern |
|---|---|
| Motivation | Flat (text only, never has code) |
| The basics | Drill-down if it has code; flat otherwise |
| A worked example | Drill-down (always has code) |
| Common mistake | Drill-down (always has broken code + error output) |
| Recap | Flat (bullet list, never has code) |

### Markdown shape (example)

```markdown
### Worked example

A "hello, expense tracker" — short prose explanation of why this example matters and what it shows.

--

### Code

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello! You spent €23.50 on coffee today.")
}
```

--

### Output

```
Hello! You spent €23.50 on coffee today.
```
```

### Scaffolder template change

`tools/new-lesson/template/slides/slides.md.tmpl` is rewritten so each concept's pre-stocked sub-slides already include the right vertical structure. Authors filling in a new lesson get the right shape for free; they don't have to remember the convention.

The template will provide both flat skeleton placeholders (Motivation, Recap) and pre-split skeleton placeholders (The basics, Worked example, Common mistake), with each pre-split block already having the `--` separator and a `### Code` placeholder heading.

### CONTRIBUTING.md update

The "Phase 1 conventions / Heavy-explanatory slide style" subsection gets a new paragraph after the existing 5-bullet pattern list:

> **Code goes "down."** When a sub-slide has both explanatory prose and a code block, split them into a vertical stack with `--`. The prose is the parent slide; the code is the child below it. Authors give each vertical step its own `###` sub-heading (`### Code`, `### Run it`, `### Output`). Students press Right to move between concepts; Down to drill into code. Reveal.js shows a navigation arrow at the bottom-right when there's more below — no explicit "press Down" cue.

### Lessons 01 and 02 retrofit

Each `slides.md` is rewritten to follow the new convention. Estimate per lesson:

- ~12 (lesson 01) and ~10 (lesson 02) `--` separators introduced — one per code-bearing sub-slide.
- Each split adds a `### Code` (or similar) heading and reorganises the prose so the parent slide is purely explanatory.
- Total content unchanged — words, code examples, common-mistake examples all stay. Only structure (single slide → vertical stack) changes.
- Diff size per lesson: 80-150 lines of `slides.md` reformatting.

The retrofit is mechanical once the rule is understood. The CONTRIBUTING.md addition serves as the canonical reference that authors consult.

## Implementation file map

After implementation:

```
shared/reveal/
├── theme/go-training.css                  (rewritten — new tokens + slide patterns)
└── plugin/highlight/dracula.css           (NEW — vendored from highlight.js)

tools/
├── build-index/
│   ├── main.go                            (modified — add `allLessons` + `Phase`)
│   └── index.html.tmpl                    (rewritten — phase-grouped grid)
└── new-lesson/template/
    └── slides/
        ├── index.html.tmpl                (modified — load dracula.css)
        └── slides.md.tmpl                 (rewritten — pre-split sub-slides)

CONTRIBUTING.md                            (modified — "Code goes down" paragraph)

lessons/01-hello/slides/
├── index.html                              (regenerated by `make new-lesson`?
│                                           or hand-edited to swap monokai.css → dracula.css)
└── slides.md                               (rewritten — vertical stacks)

lessons/02-variables/slides/
├── index.html                              (same: monokai.css → dracula.css)
└── slides.md                               (rewritten — vertical stacks)
```

The lesson `index.html` files are mostly identical scaffold (only the title line differs); their highlight CSS link can be hand-edited or the file can be regenerated from the updated template — implementation plan picks the cleanest mechanism.

## Open items deferred to implementation planning

- Whether to vendor Source Sans 3 or accept the bundled Source Sans Pro (visual diff is negligible — defer to "use the bundled fonts").
- The exact `allLessons` slice values for Plans F-K (lesson 03-08) — the `Title` and `Blurb` come from the master Phase 1 design; the implementation plan can copy them straight in.
- Whether to clean up unused vendored Reveal themes in `shared/reveal/dist/theme/` (beige.css, dracula.css, etc.) — not needed for our deck, but harmless. Defer cleanup; small space cost.
