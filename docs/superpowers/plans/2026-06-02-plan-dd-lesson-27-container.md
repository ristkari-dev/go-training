# Plan DD — Lesson 27 (Build, release & containerization) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.
>
> **Commit policy:** NO `Co-Authored-By` trailer or "Generated with" line in commit messages or PR bodies. Subject + body only.

**Goal:** Author lesson 27 — the fifth Phase 4 lesson. Make the `logstatsd` daemon shippable: link-time version stamping (`-ldflags -X`), a static `CGO_ENABLED=0` binary, a multi-stage Dockerfile on a distroless base, and image hygiene.

**Architecture:** Same per-lesson pattern, six tasks. Carries the entire L26 daemon (`logstatsd` + config + httpsrv + grpcsrv + logparse + logstats + proto/generated) verbatim. New: a `warmup/buildinfo` package (build-metadata vars + banner), a `-version` flag wired into `logstatsd` (provided), a provided `Dockerfile` + `.dockerignore` + `Makefile` build targets, and a version-stamping integration test. Five-concept slide deck.

**Tech Stack:** Go 1.23 stdlib (`fmt`, `os`, `os/exec`, `path/filepath` for the build test) + carried grpc/protobuf. Tooling: `go build -ldflags`, Docker (multi-stage, distroless). No new Go deps.

---

## Scope

After Plan DD: lesson 27 complete; `make test` + `make test-race` green; `logstatsd -version` prints stamped build metadata; the Dockerfile builds a ~10MB distroless image that runs; `27-container` in the index; go directive stays `go 1.23`.

### Design decisions (2 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **Always-run go-build test; opt-in docker test.** The version-stamp integration test does `go build -ldflags -X ...Version=test` + runs `-version` (fast, no Docker, always runs). A separate docker-build test skips unless `DOCKER_TEST=1` AND `docker` is on PATH — so normal CI stays fast and keeps the "go test from a clean checkout" ethos.
2. **Five slide concepts:** build flags + version stamping · static binaries + cross-compilation · build tags · multi-stage Dockerfile + distroless · image hygiene.

**Plan-recommended:**

3. **`buildinfo` is the warmup** — `Version`/`Commit`/`Date` link-time vars + `String()` banner. The exercise is `String()`; the lesson's real teaching (the `-ldflags` mechanism) is shown via the integration test + Dockerfile.
4. **`-version` is pre-scanned in `main`** before `config.Load`, so the config FlagSet never sees it (avoids a parse error). Provided (not the exercise).
5. **Dockerfile/.dockerignore/Make targets are provided artifacts** (Dockerfiles don't fill-in-the-blank well). The Dockerfile lives once at the lesson root and builds `./lessons/27-container/solutions/cmd/logstatsd` from the repo-root context.
6. **Carry the whole daemon forward** (keeps grpc/protobuf imported → go.mod deps + directive stable).

### Verified facts (prototyped before writing this plan)

- `buildinfo` + `String()` + unit test (defaults `dev`/`none`/`unknown`) — pass.
- The version-stamp integration test works **only when building by import path** (`go build ... <module>/.../cmd/logstatsd`), NOT `./cmd/logstatsd` (the test runs in the package dir, so a relative path double-nests). Verified: stamps `Version`+`Commit`, runs `-version`, asserts.
- A real multi-stage distroless build (`golang:1.23` builder → `CGO_ENABLED=0 go build -ldflags` → `gcr.io/distroless/static:nonroot`) builds, runs `-version`, prints the stamped version, and is **~10MB**. Verified end-to-end with local Docker.
- `ARG VERSION` → `-ldflags -X` → binary `-version` output flows correctly.

---

## File structure

```
lessons/27-container/
├── Dockerfile                                   (Task 5 — provided)
├── .dockerignore                                (Task 5 — provided)
├── README.md                                    (Task 6)
├── slides/{index.html, slides.md, assets/.gitkeep}   (Task 5/6)
├── exercises/
│   ├── warmup/buildinfo/{buildinfo.go, buildinfo_test.go}   (Task 2 — String SKELETON)
│   ├── warmup/config/                           (Task 1 — carried)
│   ├── proto/ + logstatspb/                     (Task 1 — carried, regenerated)
│   ├── internal/{logparse,logstats,grpcsrv,httpsrv}/   (Task 1 — carried)
│   └── cmd/logstatsd/{main.go, main_test.go, version_test.go}   (Task 3 — -version provided; version_test SKELETON)
└── solutions/   (mirrored; version_test real)
```

---

## Task 1: Scaffold + carry forward + regenerate proto (controller-direct)

- [ ] **Step 1:** Scaffold + remove flat stubs:
```bash
make new-lesson NAME=27-container
rm lessons/27-container/exercises/main.go lessons/27-container/exercises/main_test.go \
   lessons/27-container/exercises/warmup.go lessons/27-container/exercises/warmup_test.go \
   lessons/27-container/solutions/main.go lessons/27-container/solutions/main_test.go \
   lessons/27-container/solutions/warmup.go lessons/27-container/solutions/warmup_test.go
```

- [ ] **Step 2:** Carry forward the whole daemon (both trees), rewrite paths + header:
```bash
SRC=lessons/26-config
DST=lessons/27-container
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/proto" "$DST/$side/warmup" "$DST/$side/cmd"
  cp -R "$SRC/$side/internal/." "$DST/$side/internal/"
  cp -R "$SRC/$side/warmup/config" "$DST/$side/warmup/config"
  cp -R "$SRC/$side/cmd/logstatsd" "$DST/$side/cmd/logstatsd"
  cp "$SRC/$side/proto/logstats.proto" "$DST/$side/proto/logstats.proto"
done
grep -rl '26-config' "$DST" | while read -r f; do sed -i '' 's#lessons/26-config#lessons/27-container#g' "$f"; done
grep -rl 'lesson 26' "$DST" | while read -r f; do sed -i '' 's/lesson 26/lesson 27/g' "$f"; done
```

- [ ] **Step 3:** Add the L27 protos to the `Makefile` `proto` target (append the two `lessons/27-container/{exercises,solutions}/proto/logstats.proto` lines), then regenerate:
```bash
make proto
go mod tidy
grep '^go ' go.mod   # MUST stay go 1.23.0
```

- [ ] **Step 4:** Verify the carried baseline:
```bash
go build ./lessons/27-container/...
go test ./lessons/27-container/... 2>&1 | tail -12
go vet ./lessons/27-container/... && golangci-lint run ./lessons/27-container/... && gofmt -l lessons/27-container/
```
Expected: build clean; carried tests pass (config, httpsrv, grpcsrv, logstatsd's TestServe, logparse, logstats); vet/lint/fmt clean.

- [ ] **Step 5:** Commit:
```bash
git add lessons/27-container Makefile go.mod go.sum
git commit -m "chore(lesson-27): scaffold + carry forward the L26 logstatsd daemon"
```

---

## Task 2: Warm-up — `buildinfo`

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/27-container/{exercises,solutions}/warmup/buildinfo`

- [ ] **Step 2:** `exercises/warmup/buildinfo/buildinfo.go` (SKELETON — vars provided; `String` panics):

```go
// Package buildinfo holds build metadata stamped at link time via
// -ldflags "-X". The defaults apply to a plain `go build`/`go run`; a
// release build overrides them (see the Dockerfile and the Makefile
// `build` target).
package buildinfo

import "fmt"

var (
	Version = "dev"     // -ldflags -X .../buildinfo.Version=...
	Commit  = "none"    // -ldflags -X .../buildinfo.Commit=...
	Date    = "unknown" // -ldflags -X .../buildinfo.Date=...
)

// String renders a one-line version banner:
//
//	logstatsd <version> (commit <commit>, built <date>)
//
// Hint: fmt.Sprintf("logstatsd %s (commit %s, built %s)", Version, Commit, Date)
func String() string {
	_ = fmt.Sprintf
	panic("TODO: return a one-line banner with Version, Commit, Date")
}
```

- [ ] **Step 3:** `exercises/warmup/buildinfo/buildinfo_test.go` (SKELETON):

```go
package buildinfo

import "testing"

// TestBuildInfo is a SKELETON. Assert the defaults (dev/none/unknown)
// and that String() includes them.
func TestBuildInfo(t *testing.T) {
	// TODO: check Version=="dev", Commit=="none", Date=="unknown";
	//       assert String() contains each.
	_ = String
}
```

- [ ] **Step 4:** `solutions/warmup/buildinfo/buildinfo.go` (same vars; `String` implemented):

```go
// Package buildinfo holds build metadata stamped at link time via
// -ldflags "-X". Reference implementation.
package buildinfo

import "fmt"

var (
	Version = "dev"     // -ldflags -X .../buildinfo.Version=...
	Commit  = "none"    // -ldflags -X .../buildinfo.Commit=...
	Date    = "unknown" // -ldflags -X .../buildinfo.Date=...
)

// String renders a one-line version banner.
func String() string {
	return fmt.Sprintf("logstatsd %s (commit %s, built %s)", Version, Commit, Date)
}
```

- [ ] **Step 5:** `solutions/warmup/buildinfo/buildinfo_test.go`:

```go
package buildinfo

import (
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	if Version != "dev" || Commit != "none" || Date != "unknown" {
		t.Errorf("defaults = %q / %q / %q", Version, Commit, Date)
	}
}

func TestString(t *testing.T) {
	s := String()
	for _, want := range []string{"logstatsd", "dev", "none", "unknown"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, missing %q", s, want)
		}
	}
}
```

- [ ] **Step 6:** Verify + commit:
```bash
gofmt -l lessons/27-container/
go test ./lessons/27-container/exercises/warmup/buildinfo/... 2>&1 | tail -5
go test -v ./lessons/27-container/solutions/warmup/buildinfo/... 2>&1 | tail -10
make test && go vet ./lessons/27-container/... && golangci-lint run ./lessons/27-container/...

git add lessons/27-container/exercises/warmup/buildinfo lessons/27-container/solutions/warmup/buildinfo
git commit -m "feat(lesson-27): warmup — buildinfo (link-time version vars + banner)"
```

Expected: exercises vacuous-pass; solutions 2 tests PASS.

---

## Task 3: Wire `-version` into `logstatsd` + the version-stamp test

The carried `logstatsd/main.go` gains a `-version` pre-scan (provided). Add the version-stamping integration test (skeleton in exercises, real in solutions).

- [ ] **Step 1:** Edit BOTH `lessons/27-container/{exercises,solutions}/cmd/logstatsd/main.go`: add the `buildinfo` import (per-tree path) and a pre-scan at the very top of `main`, before `config.Load`:

```go
func main() {
	// Pre-scan for -version/--version before config parsing so it works
	// regardless of the config FlagSet.
	for _, a := range os.Args[1:] {
		if a == "-version" || a == "--version" {
			fmt.Println(buildinfo.String())
			return
		}
	}

	cfg, err := config.Load(os.Args[1:], os.Getenv)
	// ... unchanged ...
}
```

Add the import `github.com/ristkari-dev/go-training/lessons/27-container/<tree>/warmup/buildinfo` and ensure `fmt` is imported (it already is). This is PROVIDED in both trees (not the exercise) — but note: in the exercises tree `buildinfo.String()` panics, so running `logstatsd -version` there panics until the student implements it. That's fine; it's not invoked by any vacuous test.

- [ ] **Step 2:** Create `exercises/cmd/logstatsd/version_test.go` (SKELETON):

```go
package main

import "testing"

// TestVersionStamping is a SKELETON. Build the logstatsd binary with
// -ldflags -X stamping buildinfo.Version, run it with -version, and
// assert the stamped value appears in the output. (See the solution for
// the exec/go-build pattern.)
func TestVersionStamping(t *testing.T) {
	// TODO
}
```

- [ ] **Step 3:** Create `solutions/cmd/logstatsd/version_test.go` (real — VERIFIED pattern; build by IMPORT PATH, not relative):

```go
package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionStamping builds logstatsd with -ldflags -X stamping a
// version + commit into buildinfo, runs `-version`, and asserts the
// stamped values appear. Proves the link-time stamping mechanism
// end-to-end without Docker. Always runs (only needs the go toolchain).
func TestVersionStamping(t *testing.T) {
	const pkg = "github.com/ristkari-dev/go-training/lessons/27-container/solutions/cmd/logstatsd"
	const buildinfoPkg = "github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo"

	bin := filepath.Join(t.TempDir(), "logstatsd")
	ldflags := "-X " + buildinfoPkg + ".Version=test-v9 -X " + buildinfoPkg + ".Commit=abc123"
	// Build by IMPORT PATH (not ./...): the test's CWD is this package's
	// dir, so a relative path would double-nest.
	build := exec.Command("go", "build", "-ldflags", ldflags, "-o", bin, pkg)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	out, err := exec.Command(bin, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("run -version: %v\n%s", err, out)
	}
	got := string(out)
	for _, want := range []string{"test-v9", "abc123"} {
		if !strings.Contains(got, want) {
			t.Errorf("-version output %q missing %q", strings.TrimSpace(got), want)
		}
	}
}

// TestDockerBuild builds + runs the image. Opt-in: skips unless
// DOCKER_TEST=1 and docker is on PATH, so normal CI stays fast.
func TestDockerBuild(t *testing.T) {
	if os.Getenv("DOCKER_TEST") != "1" {
		t.Skip("set DOCKER_TEST=1 to run the docker build test")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not on PATH")
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "..")) // → repo root
	if err != nil {
		t.Fatal(err)
	}
	tag := "logstatsd:l27test"
	build := exec.Command("docker", "build",
		"-f", filepath.Join("lessons", "27-container", "Dockerfile"),
		"--build-arg", "VERSION=docker-v9",
		"-t", tag, ".")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("docker build: %v\n%s", err, out)
	}
	defer exec.Command("docker", "rmi", "-f", tag).Run() //nolint:errcheck

	out, err := exec.Command("docker", "run", "--rm", tag, "-version").CombinedOutput()
	if err != nil {
		t.Fatalf("docker run -version: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "docker-v9") {
		t.Errorf("docker -version output %q missing docker-v9", strings.TrimSpace(string(out)))
	}
}
```

Add `"os"` to the import block (used by `TestDockerBuild`).

- [ ] **Step 4:** Verify + commit:
```bash
gofmt -l lessons/27-container/
go test ./lessons/27-container/exercises/cmd/logstatsd/... 2>&1 | tail -5
go test -v -run 'TestVersionStamping|TestServe' ./lessons/27-container/solutions/cmd/logstatsd/... 2>&1 | tail -15
go test -race ./lessons/27-container/solutions/cmd/logstatsd/... 2>&1 | tail -3
make test && make test-race && go vet ./lessons/27-container/... && golangci-lint run ./lessons/27-container/...

git add lessons/27-container/exercises/cmd lessons/27-container/solutions/cmd
git commit -m "feat(lesson-27): logstatsd -version flag + version-stamp integration test"
```

Expected: exercises vacuous-pass; solutions TestVersionStamping + TestServe PASS, TestDockerBuild SKIP; -race clean.

> Note: `go test -race` will also run TestVersionStamping, which shells out to `go build` (no -race on the child) — fine. TestServe is the race-relevant one.

---

## Task 4: Dockerfile + .dockerignore + Makefile targets (controller-direct)

- [ ] **Step 1:** Create `lessons/27-container/Dockerfile` (VERIFIED shape, adapted to the repo-root context + real import paths):

```dockerfile
# syntax=docker/dockerfile:1

# ---- build stage: full Go toolchain ----
FROM golang:1.23 AS build
WORKDIR /src

# Cache module downloads first.
COPY go.mod go.sum ./
RUN go mod download

# Then the source (the build context is the repo root; .dockerignore trims it).
COPY . .

ARG VERSION=docker
ARG COMMIT=none
ARG DATE=unknown
ENV CGO_ENABLED=0
RUN go build \
    -ldflags "-X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Version=${VERSION} \
              -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Commit=${COMMIT} \
              -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Date=${DATE}" \
    -o /out/logstatsd ./lessons/27-container/solutions/cmd/logstatsd

# ---- final stage: tiny, non-root, no shell ----
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/logstatsd /logstatsd
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/logstatsd"]
```

- [ ] **Step 2:** Create `lessons/27-container/.dockerignore`:

```
.git
dist
**/*_test.go
docs
*.md
```

> Note: excluding `**/*_test.go` keeps tests out of the image build (they're not needed to `go build` the binary). Keep it conservative — do NOT exclude `go.sum` or any non-test `.go`.

- [ ] **Step 3:** Add `build` + `docker-build` targets to the repo `Makefile`:

```makefile
.PHONY: build
build: ## Build a version-stamped static logstatsd into bin/ (lesson 27)
	@CGO_ENABLED=0 go build \
	  -ldflags "-X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev) \
	            -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Commit=$$(git rev-parse --short HEAD 2>/dev/null || echo none) \
	            -X github.com/ristkari-dev/go-training/lessons/27-container/solutions/warmup/buildinfo.Date=$$(git log -1 --format=%cI 2>/dev/null || echo unknown)" \
	  -o bin/logstatsd ./lessons/27-container/solutions/cmd/logstatsd
	@echo "built bin/logstatsd"

.PHONY: docker-build
docker-build: ## Build the logstatsd container image (lesson 27)
	docker build -f lessons/27-container/Dockerfile \
	  --build-arg VERSION=$$(git describe --tags --always --dirty 2>/dev/null || echo dev) \
	  --build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo none) \
	  -t logstatsd:$$(git describe --tags --always --dirty 2>/dev/null || echo dev) .
```

> `git log -1 --format=%cI` (commit date) keeps the build reproducible vs a wall-clock `date`.

- [ ] **Step 4:** Verify locally (build target always; docker manually):
```bash
make build && ./bin/logstatsd -version && rm -rf bin
gofmt -l lessons/27-container/
# Manual docker (optional, requires docker):
# DOCKER_TEST=1 go test -run TestDockerBuild ./lessons/27-container/solutions/cmd/logstatsd/
```
Expected: `make build` produces `bin/logstatsd`; `-version` prints a git-derived version (not "dev" if tags/commits exist).

- [ ] **Step 5:** Commit:
```bash
git add lessons/27-container/Dockerfile lessons/27-container/.dockerignore Makefile
git commit -m "feat(lesson-27): Dockerfile (multi-stage distroless) + .dockerignore + build targets"
```

> Ensure `bin/` is gitignored (check root `.gitignore`; add `bin/` if absent — but do NOT commit `bin/`).

---

## Task 5: Slide deck — 5 concepts (controller inline)

**File:** `lessons/27-container/slides/slides.md`. Concepts:

1. **Build flags + version stamping** — `go build -ldflags "-X pkg.Var=value"` injects metadata into string vars; `-version` surfaces it; git-derived in the Makefile. *Mistake:* hardcoding version in source; a wrong `-X` path silently no-ops (var stays "dev").
2. **Static binaries + cross-compilation** — `CGO_ENABLED=0` → fully static (runs on scratch/distroless, no libc); `GOOS`/`GOARCH` cross-compile from any host. *Mistake:* CGO-enabled binary on a minimal base → runtime "not found" (missing libc/loader).
3. **Build tags** — `//go:build` constraints for platform/feature-gated files; `go build -tags`. *Mistake:* OS-specific code with no tag breaks cross-compile; tag typos silently exclude a file.
4. **Multi-stage Dockerfile + distroless** — fat builder stage, tiny final stage with only the binary; distroless/scratch bases. *Mistake:* shipping the builder (toolchain + source in the image — huge, CVE-laden).
5. **Image hygiene** — `.dockerignore`, non-root (`distroless:nonroot`), no shell, pinned bases, minimal attack surface, `EXPOSE`. *Mistake:* `COPY . .` with no `.dockerignore` (secrets/junk in layers); running as root.

- [ ] Author the deck (title-slide-grid, "What we'll cover", "The story so far" — from a daemon to a shippable image, 5 concepts, Practice, Closing thought, What we learned, Up next → L28 observability). Then:
```bash
make slides-build && grep -q "27-container" dist/index.html && echo "✓ index" && rm -rf dist
git add lessons/27-container/slides/
git commit -m "feat(lesson-27): slides — build, release & containerization (5 concepts)"
```

---

## Task 6: README + verify + final review + PR (controller)

- [ ] **Step 1:** Confirm `tools/build-index/main.go` lists L27 as `{Number:"27", Slug:"container", Title:"Containerization", Blurb:"multi-stage · distroless", Phase:4}` — slug matches; no change.

- [ ] **Step 2:** Write `lessons/27-container/README.md` (~300 lines): "What's different from L26" (same daemon, now shippable); the version-stamping mechanism (`-ldflags -X` + the `-version` flag + `make build`); the Dockerfile walkthrough (multi-stage, static, distroless, non-root) + image size (~10MB); how to build/run the container; the always-run vs opt-in (`DOCKER_TEST=1`) test split; "going further" (reproducible builds with `-trimpath` + `-buildvcs`; SBOM/`docker scout`; `ko` for daemonless image builds; a scratch base + CA certs; multi-arch `docker buildx`).

- [ ] **Step 3:** Full sweep:
```bash
make test && make test-race
go test -v ./lessons/27-container/solutions/...
go test -race ./lessons/27-container/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/27-container/ && grep '^go ' go.mod   # go 1.23.0
make build && ./bin/logstatsd -version && rm -rf bin
make slides-build && grep -q "27-container" dist/index.html && rm -rf dist
# Opt-in docker E2E (local; needs docker):
DOCKER_TEST=1 go test -run TestDockerBuild ./lessons/27-container/solutions/cmd/logstatsd/ 2>&1 | tail -3
```

- [ ] **Step 4:** Commit README:
```bash
git add lessons/27-container/README.md
git commit -m "docs(lesson-27): README — build, release & containerization self-study"
```

- [ ] **Step 5:** Dispatch `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: the `-version` pre-scan (correct, before config.Load, doesn't break normal flag parsing); the version_test building by import path (not relative); the docker test's opt-in guard + repo-root resolution; the Dockerfile (`CGO_ENABLED=0`, `-ldflags` path matches buildinfo's real path, distroless non-root, no secrets via .dockerignore); the Makefile ldflags path correctness; `bin/` gitignored; deps/go-directive unchanged; exercises skeletons vacuous. Apply fixes.

- [ ] **Step 6:** Push + open PR (no co-author trailer); watch CI green.

---

## Verification (after Task 6)

```bash
make test && make test-race
go test ./lessons/27-container/exercises/...            # vacuous-pass
go test -v ./lessons/27-container/solutions/...         # buildinfo + version-stamp + carried daemon
go test -race ./lessons/27-container/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/27-container/ && grep '^go ' go.mod    # go 1.23.0
make build && ./bin/logstatsd -version && rm -rf bin    # git-stamped version
make slides-build && grep -q "27-container" dist/index.html && rm -rf dist

# Opt-in container E2E (local; needs docker)
DOCKER_TEST=1 go test -run TestDockerBuild ./lessons/27-container/solutions/cmd/logstatsd/
# or manually:
make docker-build && docker run --rm logstatsd:$(git describe --tags --always --dirty) -version
```

## Critical file paths

To create: `lessons/27-container/` tree (Dockerfile, .dockerignore, README, slides, warmup/buildinfo, cmd/logstatsd/version_test.go, + carried daemon), mirrored exercises/solutions.
To modify: `Makefile` (proto target += L27 protos; new `build` + `docker-build` targets), `go.mod`/`go.sum` (tidy), root `.gitignore` (ensure `bin/`).
To reference: `lessons/26-config/...` (carry-forward source), `tools/build-index/main.go` (slug `container` already correct).

## Execution after approval

1. (Branch `feature/plan-dd-lesson-27-container` already created off main.)
2. Commit this plan doc.
3. Execute: Task 1 + Task 4 controller-direct (carry/regenerate; Dockerfile/Make — environment-sensitive); subagents for Tasks 2-3 (buildinfo, -version wiring + test); controller inline for Tasks 5-6 (slides, README); controller runs all verification incl. a real `make build` and an opt-in docker E2E.
4. Final code review subagent. Push + PR (no co-author trailer); watch CI green.
