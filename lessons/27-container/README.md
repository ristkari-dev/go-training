# Lesson 27: Build, release & containerization

## What you'll learn

By the end of this lesson you can:

- Stamp build metadata into a binary at link time with **`-ldflags -X`** and surface it via a `-version` flag.
- Produce **static** binaries (`CGO_ENABLED=0`) and **cross-compile** (`GOOS`/`GOARCH`).
- Gate files by platform/feature with **build tags** (`//go:build`).
- Write a **multi-stage Dockerfile** that ships a tiny **distroless** image.
- Apply **image hygiene**: `.dockerignore`, non-root, no shell, small attack surface.

This is the fifth Phase 4 lesson. The L26 daemon is unchanged; we make it **shippable** as a ~10MB container.

## What's different from L26

Nothing about the service's behavior — `logstatsd` is carried verbatim. L27 adds the *release* layer: a `buildinfo` package + `-version` flag, a multi-stage `Dockerfile` on distroless, a `.dockerignore`, and `make build` / `make docker-build` targets. The lesson's point is that a well-built Go program is shippable almost for free.

Still **stdlib-only** (plus the carried grpc/protobuf). Docker is external tooling, not a Go dependency.

## The package layout

```
lessons/27-container/
├── Dockerfile              NEW (provided): multi-stage build → distroless
├── .dockerignore           NEW (provided)
├── {exercises,solutions}/
│   ├── warmup/buildinfo/   ← you implement: String() banner; Version/Commit/Date vars provided
│   ├── warmup/config/      carried from L26
│   ├── proto/ + logstatspb/   carried
│   ├── internal/{logparse,logstats,grpcsrv,httpsrv}/   carried
│   └── cmd/logstatsd/      carried + -version (provided) + version-stamp test
```

The Dockerfile lives once at the lesson root and builds `./solutions/cmd/logstatsd` from the **repo-root** build context.

---

## Concept 1 — Build flags + version stamping

Declare string vars with safe defaults, override them at link time:

```go
package buildinfo
var (
    Version = "dev"     // -ldflags -X .../buildinfo.Version=v1.2.3
    Commit  = "none"
    Date    = "unknown"
)
```

```bash
go build -ldflags "-X github.com/you/app/.../buildinfo.Version=v1.2.3" ./cmd/logstatsd
```

`-X importpath.Var=value` sets a package-level string var at link time. Plain `go build`/`go test` keeps the defaults (so the program always works); a release build stamps real values. `logstatsd` pre-scans `os.Args` for `-version` *before* `config.Load` (so the config flag set never sees it) and prints `buildinfo.String()`. `make build` derives the values from git (`git describe`, `rev-parse`, and `git show -s --no-show-signature --format=%cI` for a reproducible commit date).

### Common mistake

A **wrong `-X` path silently no-ops** — `-X` only sets the var when the import path + name match exactly; a typo leaves the default and ships `dev` with no error. Guard it with an integration test that builds with `-ldflags`, runs `-version`, and asserts the stamped value appears. (Also: hardcoding the version in source, which drifts.)

> Repo gotcha: this repo signs commits with `log.showSignature=true`, so `git log --format=%cI` leaks `gpg:` lines into the value. Use `git show -s --no-show-signature` (the Makefile does).

---

## Concept 2 — Static binaries + cross-compilation

```bash
CGO_ENABLED=0 go build ./cmd/logstatsd            # fully static, no libc
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build    # cross-compile
```

`CGO_ENABLED=0` disables cgo so the binary links nothing dynamically and runs on an empty base (`scratch`/`distroless`). `GOOS`/`GOARCH` target any platform from any host — no cross-toolchain (`go tool dist list` shows all combos). The Dockerfile sets `ENV CGO_ENABLED=0`; the resulting image is ~10MB.

### Common mistake

A **CGO-enabled binary on a minimal base** crashes at startup — `no such file or directory` (missing dynamic loader/libc) or silent DNS failures. Use `CGO_ENABLED=0` for `scratch`/`distroless`, or a libc base (`distroless/base`, `alpine`) if you truly need cgo.

---

## Concept 3 — Build tags

A `//go:build` line (before `package`, followed by a blank line) gates a file's compilation:

```go
//go:build linux

package platform
```

```bash
go build -tags=integration ./...
```

Filename suffixes are implicit constraints (`foo_linux.go`, `foo_arm64.go`); custom tags opt files in via `-tags`; constraints combine (`//go:build linux && amd64`, `//go:build !windows`). Cross-compilation selects the right `_GOOS` file automatically — no runtime branching.

### Common mistake

**Platform-specific code with no constraint** breaks cross-compile (compiles on your host, fails under `GOOS=darwin`). And a **malformed/typo'd `//go:build` line** (or a missing blank line after it) is silently ignored, so the file compiles always or never — not as intended. `go vet` catches some of these.

---

## Concept 4 — Multi-stage Dockerfile + distroless

```dockerfile
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download             # cached unless deps change
COPY . .
ARG VERSION=docker
ENV CGO_ENABLED=0
RUN go build -ldflags "-X .../buildinfo.Version=${VERSION}" -o /out/logstatsd ./lessons/27-container/solutions/cmd/logstatsd

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/logstatsd /logstatsd
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/logstatsd"]
```

The fat builder stage compiles; `COPY --from=build` pulls only the binary into the tiny final stage. `distroless/static` is CA certs + tzdata + a non-root user — no shell, no package manager. Copy `go.mod`/`go.sum` and `go mod download` before the source for layer caching. `ARG VERSION` threads the version into `-ldflags`; `make docker-build` passes `--build-arg VERSION=$(git describe ...)`.

### Common mistake

**Shipping the build stage** — a single-stage `FROM golang:1.23` image carries the whole toolchain *and your source* into production: hundreds of MB and a huge CVE surface. Multi-stage + a minimal final base fixes it.

---

## Concept 5 — Image hygiene

- **`.dockerignore`** keeps junk/secrets out of the build context (which `COPY . .` bakes into layers): `.git`, `dist/`, `bin/`, test files, docs.
- **Non-root** (`distroless:nonroot` + `USER nonroot:nonroot`) — a compromised process isn't root.
- **No shell** (distroless has no `/bin/sh`) — an RCE can't `sh -c`.
- **Pin base digests** for reproducibility; rebuild to pick up patches.

Our image's entire contents are CA certs and one static binary — nothing to `apt upgrade`, little to attack.

### Common mistake

**`COPY . .` with no `.dockerignore`** drags `.git` (history + possible secrets), `.env`, and build artifacts into layers — extractable even if a later layer "deletes" them. And **running as root** (the default) means a container escape starts as root. Add a `.dockerignore`; run non-root.

---

## Exercise: warm-up — `buildinfo`

Implement `String()` in `exercises/warmup/buildinfo/buildinfo.go` — the one-line banner from `Version`/`Commit`/`Date` (the vars `-ldflags -X` overrides at build time).

**Time:** 5-10 minutes.

## Exercise: main — study the release path

The version-stamping path is provided end-to-end; study how it fits:
- `cmd/logstatsd` pre-scans `-version` → `buildinfo.String()`.
- `version_test.go` builds with `-ldflags -X` (by import path!) and asserts the stamp — always runs, no Docker.
- the `Dockerfile` (multi-stage, distroless) + `make build`/`make docker-build` produce the stamped binary/image.

**Time:** 30-45 minutes (mostly building + running the container).

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/27-container/exercises
go test ./...

# Reference solution
go test -v ./lessons/27-container/solutions/...
make test-race

# Build a version-stamped static binary
make build && ./bin/logstatsd -version
# logstatsd <git-version> (commit <sha>, built <commit-date>)
rm -rf bin

# Build + run the container image (~10MB distroless)
make docker-build
docker run --rm logstatsd:$(git describe --tags --always --dirty 2>/dev/null || echo dev) -version

# The opt-in container E2E test (skips unless DOCKER_TEST=1 + docker present)
DOCKER_TEST=1 go test -run TestDockerBuild ./lessons/27-container/solutions/cmd/logstatsd/
```

## Going further

- **Reproducible builds** — add `-trimpath` and pin `-buildvcs`; compare two builds' hashes.
- **Multi-arch images** — `docker buildx build --platform linux/amd64,linux/arm64` (Go cross-compiles each).
- **`ko`** — build a Go container image with no Dockerfile (`ko build ./cmd/logstatsd`); compare size/speed.
- **SBOM + scan** — `docker scout cves` or `trivy image` the result; see how little a distroless image reports.
- **`scratch` base** — drop to `FROM scratch` and add only `ca-certificates.crt` + a non-root `USER`; see what breaks vs distroless.

---

> Phase 4 nears the end. Next: Lesson 28 — **Observability**. We instrument `logstatsd` with OpenTelemetry — spans, metrics (request/ingest counters, latency histograms), trace/log correlation — exported to stdout so it runs with no collector. The course's second (and final) deliberate third-party dependency, after gRPC.
