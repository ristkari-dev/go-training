<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">26</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>Config, secrets &amp; graceful shutdown</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Turn the service into a real daemon. Learn config precedence (flags &gt; env &gt; file &gt; defaults), the <code>flag</code> package, stdlib config files, keeping secrets out of logs, and graceful shutdown — draining in-flight HTTP requests and gRPC calls on SIGTERM. One <code>logstatsd</code> process now serves both faces.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Config precedence** — defaults < file < env < flags, and why that order.
- **The `flag` package + env** — `FlagSet`, knowing what was *explicitly set*, testable loading.
- **Config files** — JSON into a struct with the standard library (no Viper).
- **Secrets hygiene** — keep tokens out of flags and logs; redaction.
- **Graceful shutdown** — drain HTTP + gRPC on SIGINT/SIGTERM; finish in-flight, reject new.

---

## The story so far — from binaries to a daemon

L23-L25 built the service's faces: HTTP (L23/24) and gRPC (L25), each with its own little `main`. **L26 consolidates them into one daemon** — `logstatsd` — serving both faces over the same `logstats.Store`, configured once, shut down cleanly once.

That's the leap from "runs on my machine" to "runs in production." A production process must take its configuration from its environment (not hardcoded constants), never leak secrets into logs, and — critically — **stop gracefully**: when the orchestrator sends SIGTERM (a deploy, a scale-down, a node drain), finish the requests you're handling, refuse new ones, and exit zero. Drop those in-flight requests and you've turned every deploy into a handful of 500s for real users.

We extracted the HTTP router into `internal/httpsrv` (parallel to `internal/grpcsrv`) so one `run` function can mount both. The transport faces are thin; the daemon wires them to the shared store and coordinates their lifecycle.

---

## Concept 1: Config precedence

### Motivation

The same binary runs on your laptop, in CI, in staging, in prod — with different addresses, log levels, and secrets each time. Hardcoding is out. You need *layers*: sensible defaults, overridable by a config file, overridable by environment variables, overridable by command-line flags. The order matters, and it's nearly universal.

---

### The basics

**defaults < file < env < flags** — each layer overrides the ones before it.

- **Defaults** — baked-in, so the binary runs with zero config.
- **File** — a JSON/YAML/TOML file checked into config management (per-environment).
- **Env** — the 12-factor standard; how containers and orchestrators inject config.
- **Flags** — the most explicit, highest priority: an operator typing `-log-level=debug` to debug *right now* must win over everything.

```go
cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // 1. defaults
// 2. overlay JSON file (if -config / LOGSTATS_CONFIG given)
// 3. overlay env (LOGSTATS_HTTP_ADDR, ...)
// 4. overlay explicitly-set flags
```

---

### A worked example

`config.Load(args, getenv)` builds the layers in order and returns the resolved `Config`. The key subtlety is "explicitly set": a flag left at its default must NOT clobber an env or file value. We'll solve that in Concept 2.

`getenv` is **injected** (`func(string) string`) rather than calling `os.Getenv` directly — so tests pass a fake environment and never touch the real one. The daemon's `main` passes `os.Getenv`; tests pass a map lookup.

---

### Common mistake

**No precedence at all** — reading only flags, or only env, so you can't layer. The other classic: **precedence inversion**, where a config file silently overrides an explicit `-flag` an operator typed. When someone runs `-log-level=debug` to chase a fire and the file's `"info"` wins, you've hidden the very logs they need. Flags are the operator's direct intent; they win.

---

### Recap

- Layer config: **defaults < file < env < flags**.
- Defaults make the binary run with zero config; flags are the highest-priority override.
- Inject `getenv` for testable loading.

---

## Concept 2: The `flag` package + env

### Motivation

Go's standard `flag` package is enough for real CLIs — no Cobra needed for a daemon. The trick that makes precedence work is knowing *which flags the user actually typed* versus which are sitting at their default value.

---

### The basics

Use a `flag.FlagSet` (not the package-global `flag`) so loading is self-contained and testable:

```go
fs := flag.NewFlagSet("logstatsd", flag.ContinueOnError)
httpAddr := fs.String("http-addr", "", "HTTP listen address")
logLevel := fs.String("log-level", "", "log level")
fs.Parse(args)
```

Then — the crux — `fs.Visit` iterates **only the flags that were explicitly set**, while `fs.VisitAll` iterates all of them:

```go
fs.Visit(func(f *flag.Flag) {       // only set flags
    switch f.Name {
    case "http-addr": cfg.HTTPAddr = *httpAddr
    case "log-level": cfg.LogLevel = *logLevel
    }
})
```

So a `-http-addr` the user typed overrides env/file; a `-http-addr` they *didn't* type leaves the env/file value intact. That's how flags sit correctly at the top of the precedence stack.

---

### A worked example

The loader applies env overrides first, then `fs.Visit` for flags last:

```go
if v := getenv("LOGSTATS_HTTP_ADDR"); v != "" { cfg.HTTPAddr = v } // env layer
fs.Visit(func(f *flag.Flag) { /* explicitly-set flags win */ })     // flag layer
```

Run with `LOGSTATS_HTTP_ADDR=:7000` and no flag → `:7000`. Add `-http-addr=:6000` → `:6000`. The flag was *set*, so it wins; the env still beats the default.

---

### Common mistake

**Using the package-global `flag`** (`flag.String`, `flag.Parse`) — it reads `os.Args` and panics on re-parse, so it's untestable and leaks state between tests. Use a fresh `FlagSet`. The subtler bug: **treating a defaulted flag as "set"** — if you read `*httpAddr` unconditionally (not via `Visit`), an empty default clobbers the env/file value, inverting precedence.

---

### Recap

- Use `flag.NewFlagSet`, not the global `flag` — self-contained + testable.
- `fs.Visit` = only explicitly-set flags; `fs.VisitAll` = all. The distinction is what makes flags the top precedence layer.
- Inject `getenv`; don't read `os.Args`/`os.Getenv` deep in the loader.

---

## Concept 3: Config files (JSON, stdlib)

### Motivation

Flags and env are great for a handful of values; a richer config (many fields, nested structure, per-environment files in git) wants a file. You don't need Viper or a YAML dependency — `encoding/json` unmarshalling into your `Config` struct is plenty, and stays inside the course's stdlib-first ethos.

---

### The basics

```go
type Config struct {
    HTTPAddr string `json:"http_addr"`
    GRPCAddr string `json:"grpc_addr"`
    LogLevel string `json:"log_level"`
}

if path != "" {
    data, err := os.ReadFile(path)
    if err != nil { return Config{}, fmt.Errorf("config: read %s: %w", path, err) }
    if err := json.Unmarshal(data, &cfg); err != nil {
        return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
    }
}
```

`json.Unmarshal` into the *already-defaulted* `cfg` only overwrites fields present in the file — absent fields keep their defaults. The file path itself comes from a flag (`-config`) or env (`LOGSTATS_CONFIG`): config is bootstrapped by config.

---

### A worked example

The file layer sits between defaults and env. A file with `{"log_level":"warn"}` overrides the default `"info"`, but a later `LOGSTATS_LOG_LEVEL=error` or `-log-level=error` still wins. Each layer overlays the previous, field by field.

---

### Common mistake

**Letting the file override explicit flags** (loading the file *after* applying flags) — precedence inversion again. Apply the file early (right after defaults), then env, then flags last. Also: failing to wrap file errors with the path (`%w` + filename) — "no such file" with no name is a miserable thing to debug at 3am.

---

### Recap

- `encoding/json` into the defaulted struct is enough — no third-party config lib.
- Unmarshalling overlays only present fields; absent ones keep defaults.
- Apply the file layer early (defaults → **file** → env → flags); wrap errors with the path.

---

## Concept 4: Secrets hygiene

### Motivation

Configuration includes secrets — API tokens, DB passwords, signing keys. They need the same precedence machinery but extra care: a secret that lands in a flag shows up in `ps` and shell history; a secret that lands in a log line is now in your log aggregator forever, searchable by anyone with read access. Most breaches are boring: a token in a log.

---

### The basics

Two rules:

1. **Secrets come from env or a file, never a flag.** Flags are visible to every process via `ps aux` and saved in shell history. So `AuthToken` is read from `LOGSTATS_AUTH_TOKEN` only — there is no `-auth-token` flag.
2. **Never log the secret.** The danger is logging the *whole config struct* for debugging and leaking the token with it. Go's `slog` gives a clean defense: implement `slog.LogValuer` to control how your type appears in logs.

```go
func (c Config) LogValue() slog.Value {
    token := ""
    if c.AuthToken != "" { token = "***" }   // redacted
    return slog.GroupValue(
        slog.String("http_addr", c.HTTPAddr),
        slog.String("grpc_addr", c.GRPCAddr),
        slog.String("log_level", c.LogLevel),
        slog.String("auth_token", token),     // never the real value
    )
}
```

Now `logger.Info("starting", "config", cfg)` prints `auth_token:"***"` — the secret can't leak through this path even if every startup logs the config.

---

### A worked example

`logstatsd` logs its resolved config on startup: `logger.Info("starting logstatsd", "config", cfg)`. Because `Config` implements `LogValuer`, the output carries `"auth_token":"***"` — operators see that a token *is* set (useful) without seeing its value (dangerous). The redaction test asserts the raw token never appears in the log output.

---

### Common mistake

**Logging the whole config struct** without a `LogValuer` — `%+v` or a plain `slog` attr dumps the token verbatim. The other one: **secrets in flags or committed config files** — both end up somewhere searchable (process list, git history). Secrets live in the environment (or a mounted secret file), and only ever appear redacted in logs.

---

### Recap

- Secrets from env/file, **never flags** (`ps` + shell-history leak).
- **Never log secrets** — implement `slog.LogValuer` to redact (`"***"`).
- Don't commit secrets; the only place the value lives is the environment.

---

## Concept 5: Graceful shutdown

### Motivation

When Kubernetes (or `docker stop`, or Ctrl-C) wants your process gone, it sends **SIGTERM**, waits a grace period, then SIGKILLs. If you exit immediately, every in-flight request — a half-written HTTP response, a streaming gRPC ingest — is dropped: 500s for users mid-deploy. Graceful shutdown means: stop accepting new work, let in-flight work finish (within a timeout), then exit.

---

### The basics

`signal.NotifyContext` turns signals into a cancelled context (Go 1.16+):

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

Each server has a graceful-stop call:

- **HTTP:** `srv.Shutdown(ctx)` — stops accepting, waits for active handlers, honoring a timeout ctx.
- **gRPC:** `srv.GracefulStop()` — stops accepting, waits for active RPCs (including streams) to finish.

The daemon runs both and drains both when `ctx` fires:

```go
select {
case <-ctx.Done():                // SIGINT/SIGTERM
    shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    httpSrv.Shutdown(shutCtx)      // drain HTTP
    grpcSrv.GracefulStop()         // drain gRPC
    return nil
case err := <-errCh:              // a server failed first → tear down the other
    ...
}
```

The bounded `shutCtx` matters: a stuck connection can't make you hang past the orchestrator's grace period (or you get SIGKILLed mid-drain anyway).

---

### A worked example

`logstatsd`'s `serve` runs the HTTP and gRPC servers in goroutines (each reporting fatal errors on a buffered `errCh`), then `select`s on `ctx.Done()` vs `errCh`. On signal: drain both with a 10s budget, return `nil` (a requested shutdown isn't a failure). If one server crashes first: tear down the other and return the error. `http.Server.Serve` returns `ErrServerClosed` after `Shutdown` (swallowed); `grpc.Server.Serve` returns `nil` after `GracefulStop`. Tested with two ephemeral listeners: hit both faces, cancel, assert `serve` returns nil and both stopped.

---

### Common mistake

**`os.Exit` on signal** (or just letting `main` return) — kills in-flight requests instantly; that's the deploy-time 500 storm. The other: **no shutdown timeout** — `Shutdown(context.Background())` with no deadline waits forever on one stuck client, so the orchestrator SIGKILLs you anyway and you've drained nothing. Always bound the drain.

---

### Recap

- SIGTERM = "wind down"; `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`.
- Drain HTTP (`Shutdown`) + gRPC (`GracefulStop`) — finish in-flight, reject new.
- Bound the drain with a timeout; a requested shutdown returns `nil`, not an error.

---

## Practice

### Warm-up

In `exercises/warmup/config/`, implement `Load(args, getenv) (Config, error)`: defaults < JSON file < env < explicitly-set flags (use `fs.Visit`!), validate the log level. The `Config` type + `LogValue` redaction are provided.

```bash
cd lessons/26-config/exercises
go test ./warmup/config/...
```

### Main

In `cmd/logstatsd/`, implement `serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis)`: run both servers concurrently, drain both on `ctx.Done()` (`http.Shutdown` + `grpc.GracefulStop`, bounded), tear down the other if one errors first. (`run`/`main`/`newLogger` are provided.)

```bash
cd lessons/26-config/exercises
go test ./...
make test-race

# run the daemon (config via flag/env/file), then SIGTERM it
LOGSTATS_LOG_LEVEL=debug go run ./cmd/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s 127.0.0.1:8080/stats; echo
kill -TERM %1   # watch it drain cleanly
```

---

## Closing thought

Nothing here needed a framework: `flag`, `encoding/json`, `os/signal`, and `slog` — all standard library — gave us layered config, file loading, secret redaction, and coordinated shutdown of two servers. Config libraries (Viper) and process managers exist, but for a single daemon the stdlib is the right amount of machinery.

The throughline of this lesson is **operability**: a service isn't done when it handles requests; it's done when an operator can configure it without recompiling, trust it not to leak secrets, and deploy a new version without dropping traffic. Those are the unglamorous properties that decide whether something is pleasant to run at 3am — and they're as much a part of the design as the API.

---

## What we learned

- **Precedence**: defaults < file < env < flags; flags are the operator's highest-priority intent.
- **`flag` package**: a `FlagSet` + `fs.Visit` (only explicitly-set flags) makes precedence correct and loading testable; inject `getenv`.
- **Config files**: `encoding/json` into the defaulted struct — overlays present fields only; no third-party lib.
- **Secrets**: env/file not flags; never log them; `slog.LogValuer` redaction.
- **Graceful shutdown**: `signal.NotifyContext` + bounded `http.Shutdown` + `grpc.GracefulStop`; finish in-flight, return nil.

---

## Up next

Lesson 27 — **Build, release & containerization**. We ship `logstatsd`: version-stamping via `-ldflags`, static `CGO_ENABLED=0` cross-compiled binaries, a multi-stage Dockerfile on a distroless base, and image hygiene. The daemon we just made configurable and gracefully-shutting-down is exactly what a container orchestrator wants to run.
