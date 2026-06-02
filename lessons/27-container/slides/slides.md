<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">27</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>Build, release &amp; containerization</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Ship the daemon. Learn link-time version stamping (<code>-ldflags -X</code>), static <code>CGO_ENABLED=0</code> binaries and cross-compilation, build tags, and a multi-stage Dockerfile on a distroless base — a ~10MB, non-root, shell-less image of the <code>logstatsd</code> service.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Build flags + version stamping** — inject version/commit/date at link time.
- **Static binaries + cross-compilation** — `CGO_ENABLED=0`, `GOOS`/`GOARCH`.
- **Build tags** — compile files conditionally.
- **Multi-stage Dockerfile + distroless** — a tiny image with just the binary.
- **Image hygiene** — `.dockerignore`, non-root, no shell, small attack surface.

---

## The story so far — from a daemon to a shippable artifact

L26 made `logstatsd` a configurable, gracefully-shutting-down daemon. It runs great — *on your machine*. L27 makes it **shippable**: a single static binary stamped with its version, baked into a minimal container image an orchestrator can pull and run anywhere.

This is where Go's deployment story shines. A Go program compiles to one self-contained binary — no interpreter, no virtualenv, no `node_modules`. With `CGO_ENABLED=0` it's *fully static*: it depends on nothing in the base image, so it runs on `scratch` (literally empty) or `distroless` (just CA certs + a non-root user, no shell, no package manager). The result is a container that's tiny, fast to pull, and has almost no attack surface — the opposite of a 1GB image with a full OS you have to keep patching.

Nothing here is Go-specific magic; it's disciplined build tooling. Let's build it up.

---

## Concept 1: Build flags + version stamping

### Motivation

A running binary should be able to tell you exactly what it is: which version, which commit, built when. You could hardcode a `const Version = "1.2.3"` — but it drifts the moment you forget to bump it, and it can't know the git commit. The fix is to inject build metadata *at link time*, from your build system, into the binary.

---

### The basics

Declare plain string vars with safe defaults, then override them with `-ldflags -X`:

```go
package buildinfo

var (
    Version = "dev"     // -ldflags -X .../buildinfo.Version=v1.2.3
    Commit  = "none"
    Date    = "unknown"
)
```

```bash
go build -ldflags "-X github.com/you/app/buildinfo.Version=v1.2.3" ./cmd/app
```

`-X importpath.Var=value` sets a package-level string variable at link time. A plain `go build`/`go run` (or `go test`) leaves the defaults, so the program always works; a release build stamps real values. Surface them with a `-version` flag.

---

### A worked example

`logstatsd` pre-scans for `-version` before parsing config (so the config flag set never chokes on it) and prints the banner:

```go
for _, a := range os.Args[1:] {
    if a == "-version" || a == "--version" {
        fmt.Println(buildinfo.String())  // "logstatsd v1.2.3 (commit abc123, built 2026-...)"
        return
    }
}
```

The `make build` target derives the values from git:

```make
LDFLAGS := -X .../buildinfo.Version=$(shell git describe --tags --always --dirty) \
           -X .../buildinfo.Commit=$(shell git rev-parse --short HEAD) \
           -X .../buildinfo.Date=$(shell git show -s --no-show-signature --format=%cI HEAD)
```

(Commit *date* via git, not wall-clock `date`, keeps the build reproducible. And `--no-show-signature` matters if your repo signs commits — otherwise git leaks signature lines into the value.)

---

### Common mistake

**A wrong `-X` path silently no-ops.** `-X` only sets the var if the import path + name match *exactly*; a typo (wrong package path, or `version` vs `Version`) leaves the default in place with no error. Your release ships reporting `dev`. The fix: an integration test that builds with `-ldflags`, runs `-version`, and asserts the stamped value actually appears — proving the path is right. (And: hardcoding the version in source, which drifts.)

---

### Recap

- Declare `var Version = "dev"`; override with `go build -ldflags "-X path.Version=..."`.
- Surface it via a `-version` flag; derive values from git in your build target.
- A wrong `-X` path no-ops silently — test that stamping actually works.

---

## Concept 2: Static binaries + cross-compilation

### Motivation

To run on `scratch`/`distroless` — images with no libc, no dynamic loader — your binary must be *fully static*: no runtime dependency on shared libraries. And to build a Linux container from a Mac (or an arm64 image from x86), you must cross-compile. Go makes both a matter of environment variables.

---

### The basics

```bash
CGO_ENABLED=0 go build ./cmd/logstatsd          # fully static (no libc dependency)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build  # cross-compile for linux/arm64
```

- **`CGO_ENABLED=0`** disables cgo, so the binary uses Go's pure-Go implementations (net, DNS) and links *nothing* dynamically. It then runs on an empty base image.
- **`GOOS`/`GOARCH`** target any OS/architecture from any host — no cross-toolchain to install. `go tool dist list` shows all combos.

By default `go build` on Linux *may* link libc dynamically (via cgo, e.g. for DNS). Setting `CGO_ENABLED=0` guarantees a static binary.

---

### A worked example

The Dockerfile's build stage sets `ENV CGO_ENABLED=0`, producing a static `logstatsd` that the final `distroless/static` stage runs with no libc present. The same binary would run on `scratch`. Verified: the resulting image is ~10MB — almost entirely the binary itself.

---

### Common mistake

**A CGO-enabled binary on a minimal base.** Build with the default `CGO_ENABLED=1` (or use a cgo-dependent library), drop it on `scratch`/`distroless`, and it crashes at startup — `no such file or directory` (the dynamic loader/libc it needs isn't there) or DNS resolution silently failing. Always `CGO_ENABLED=0` for minimal-base images (or use a base with libc, like `distroless/base` or `alpine`).

---

### Recap

- `CGO_ENABLED=0` → fully static binary that runs on `scratch`/`distroless`.
- `GOOS`/`GOARCH` cross-compile from any host with no extra toolchain.
- CGO + minimal base = runtime "not found" crash; disable cgo or use a libc base.

---

## Concept 3: Build tags

### Motivation

Sometimes a file should compile only on certain platforms, or only for certain builds (debug vs release, an enterprise feature, an integration-test-only helper). Build tags (build constraints) gate compilation at the file level.

---

### The basics

A `//go:build` line at the top of a file (before `package`) controls when it's compiled:

```go
//go:build linux

package platform
// ... linux-only implementation ...
```

```bash
go build -tags=integration ./...   # also compiles files marked //go:build integration
```

- File-name suffixes are an implicit form: `foo_linux.go`, `foo_arm64.go`, `foo_windows.go` compile only on that GOOS/GOARCH.
- Custom tags (`//go:build integration`) let you opt files in via `-tags`.
- Constraints combine: `//go:build linux && amd64`, `//go:build !windows`.

---

### A worked example

Cross-compilation and build tags work together: if you have an OS-specific implementation (say, a `reload_unix.go` using SIGHUP and a `reload_windows.go` no-op), the right one is selected automatically by GOOS when you cross-compile — no `#ifdef`, no runtime branching. Each file is plain Go; the constraint decides whether it's in the build.

---

### Common mistake

**OS-specific code with no constraint** breaks cross-compilation: a file calling a Linux-only syscall compiles fine on your Linux host but fails when you `GOOS=darwin go build`. Gate it with a build tag (or a `_linux.go` suffix). The subtle one: a **malformed/typo'd `//go:build` line** (or missing the blank line after it) is silently ignored — the file compiles always (or never), not as you intended. `go vet` catches some of these.

---

### Recap

- `//go:build <constraint>` gates a file's compilation; `_GOOS`/`_GOARCH` suffixes do it implicitly.
- `-tags=name` opts custom-tagged files into the build.
- Untagged platform-specific code breaks cross-compile; a malformed tag is silently ignored.

---

## Concept 4: Multi-stage Dockerfile + distroless

### Motivation

Building inside Docker needs the full Go toolchain (~1GB) and your source. *Running* needs only the compiled binary. A multi-stage build uses a fat "builder" stage to compile, then copies just the binary into a tiny final stage — so the toolchain and source never ship.

---

### The basics

```dockerfile
# build stage: full toolchain
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download              # cached unless deps change
COPY . .
ARG VERSION=docker
ENV CGO_ENABLED=0
RUN go build -ldflags "-X .../buildinfo.Version=${VERSION}" -o /out/logstatsd ./cmd/logstatsd

# final stage: just the binary
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/logstatsd /logstatsd
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/logstatsd"]
```

- **`AS build`** names the builder stage; **`COPY --from=build`** pulls only the binary out.
- **`distroless/static`** is just CA certs + tzdata + a non-root user — no shell, no apt, no busybox. Pair it with a static (`CGO_ENABLED=0`) binary.
- **Layer caching:** copy `go.mod`/`go.sum` and `go mod download` *before* the source, so dependency downloads are cached across source changes.
- **`ARG VERSION`** threads the version into the build's `-ldflags`.

---

### A worked example

`make docker-build` runs `docker build` with `--build-arg VERSION=$(git describe ...)`, so the image's `logstatsd -version` reports the real git version. The two stages mean the final image (~10MB) contains the binary and almost nothing else — not the 1GB of Go toolchain that built it.

---

### Common mistake

**Shipping the build stage** — a single-stage `FROM golang:1.23` image carries the entire toolchain *and your source code* into production: hundreds of MB, slow to pull, and a huge CVE surface (every tool in the image is something to patch). Multi-stage + a minimal final base is the fix.

---

### Recap

- Multi-stage: fat builder compiles; tiny final stage gets only the binary via `COPY --from`.
- `distroless/static` + a `CGO_ENABLED=0` binary = ~10MB, no shell, non-root.
- Copy `go.mod`/`go.sum` + `go mod download` before source for layer caching.

---

## Concept 5: Image hygiene

### Motivation

A container image is an artifact you operate for years. Bloat slows every pull; junk in layers leaks secrets; running as root and shipping a shell hand attackers tools. Good hygiene makes images small, safe, and boring.

---

### The basics

- **`.dockerignore`** — keep junk and secrets out of the build context (which `COPY . .` would otherwise bake in): `.git`, `dist/`, test files, docs, `.env`. Smaller context = faster builds + no accidental secret in a layer.
- **Non-root** — `distroless:nonroot` + `USER nonroot:nonroot`; a compromised process isn't root.
- **No shell** — distroless has no `/bin/sh`, so an RCE can't `sh -c`. (Debug with an ephemeral container or `distroless:debug` when needed.)
- **Pin bases** — pin a digest (`golang:1.23@sha256:...`) for reproducibility, and rebuild to pick up base patches.
- **`EXPOSE`** documents ports; it's metadata, not enforcement.

---

### A worked example

Our `.dockerignore` trims `.git`, `dist`, `bin`, `**/*_test.go`, `docs`, and `*.md` — the build context is just what `go build` needs. The final image runs as `nonroot` with no shell. The result: a small image whose entire contents are CA certs and one static binary — nothing to `apt upgrade`, nothing for an attacker to pivot through.

---

### Common mistake

**`COPY . .` with no `.dockerignore`** drags everything — `.git` (full history, possibly secrets), local `.env` files, build artifacts — into the image layers, where it's extractable even if a later layer "deletes" it. And **running as root** (the default) means a container escape starts with root. Add a `.dockerignore`; run non-root.

---

### Recap

- `.dockerignore` keeps secrets/junk out of context and layers.
- Run **non-root**, ship **no shell**, **pin** base digests, rebuild for patches.
- A clean image is CA certs + your static binary — and little else to attack.

---

## Practice

### Warm-up

In `exercises/warmup/buildinfo/`, implement `String()` — the one-line version banner from the `Version`/`Commit`/`Date` vars. (Those vars are what `-ldflags -X` overrides at build time.)

```bash
cd lessons/27-container/exercises
go test ./warmup/buildinfo/...
```

### Main

Study the version-stamping path end to end:
- `cmd/logstatsd` pre-scans `-version` and prints `buildinfo.String()`.
- The **version-stamp test** builds the binary with `-ldflags -X` and asserts the stamped value appears (always runs; no Docker).
- The **Dockerfile** (multi-stage, distroless) + `make build` / `make docker-build` produce the stamped binary/image.

```bash
make build && ./bin/logstatsd -version       # git-stamped static binary
make docker-build                            # ~10MB distroless image
docker run --rm logstatsd:$(git describe --tags --always --dirty) -version

# the opt-in container E2E test:
DOCKER_TEST=1 go test -run TestDockerBuild ./lessons/27-container/solutions/cmd/logstatsd/
```

---

## Closing thought

Deployment is where a lot of Go's design pays off at once: one static binary, no runtime, cross-compiles anywhere, drops into an empty image. The tooling here — `-ldflags`, `CGO_ENABLED=0`, multi-stage builds, distroless — isn't exotic; it's the boring, repeatable recipe that turns "it runs" into "it ships, reproducibly, as a 10MB artifact I can run anywhere and trust."

And notice what *didn't* change: the daemon's code. We added a `buildinfo` package and a `-version` flag, and wrote a Dockerfile — but the service from L26 is untouched. A well-built program is shippable almost for free; the container is a thin wrapper around a binary that was already self-contained.

---

## What we learned

- **Version stamping**: `-ldflags -X path.Var=value` injects build metadata; surface it via `-version`; test that the stamp lands.
- **Static + cross-compile**: `CGO_ENABLED=0` → runs on scratch/distroless; `GOOS`/`GOARCH` build anywhere.
- **Build tags**: `//go:build` gates files by platform/feature; untagged platform code breaks cross-compile.
- **Multi-stage + distroless**: fat builder, tiny final image (~10MB) with only the binary; cache deps before source.
- **Image hygiene**: `.dockerignore`, non-root, no shell, pinned bases — small and low-attack-surface.

---

## Up next

Lesson 28 — **Observability**. We instrument `logstatsd` with OpenTelemetry: spans across requests, metrics (request/ingest counters, latency histograms), and trace/log correlation — exported to stdout so the lesson runs with no collector. The second (and last) deliberate third-party dependency of the course, after gRPC.
