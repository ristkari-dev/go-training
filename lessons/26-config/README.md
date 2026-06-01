# Lesson 26: Configuration, secrets & graceful shutdown

## What you'll learn

By the end of this lesson you can:

- Resolve configuration with **precedence**: defaults < JSON file < env < explicitly-set flags.
- Use the **`flag` package** correctly (a `FlagSet` + `fs.Visit`) for testable, layered loading.
- Load **config files** with the standard library (`encoding/json`) — no Viper.
- Keep **secrets** out of flags and logs (`slog.LogValuer` redaction).
- Shut down **gracefully** — drain in-flight HTTP requests and gRPC calls on SIGINT/SIGTERM.

This is the fourth Phase 4 lesson. The three separate L25 binaries become one configurable, gracefully-shutting-down daemon: **`logstatsd`**.

## What's different from L25

L23-L25 gave the service an HTTP face and a gRPC face, each with its own `main`. **L26 consolidates them into one `logstatsd` daemon** serving both over the same `logstats.Store`. To make that possible, the HTTP router/handlers/middleware were extracted from the old `cmd/logstats-server` into **`internal/httpsrv`** (parallel to `internal/grpcsrv`), so one `run` function mounts both faces. The three L25 cmd binaries are retired.

Still **stdlib-only** except the carried grpc/protobuf deps (the gRPC face rides along, which also keeps those deps live in `go.mod`).

## The package layout

```
lessons/26-config/{exercises,solutions}/
├── warmup/config/          ← you implement: Load(args, getenv) + Config.LogValue() redaction
├── proto/logstats.proto + logstatspb/   carried from L25 (generated, committed)
├── internal/
│   ├── logparse/           carried
│   ├── logstats/           carried (the shared Store)
│   ├── grpcsrv/            carried from L25 (the gRPC face)
│   └── httpsrv/            NEW (provided): the HTTP face, extracted so the daemon can mount it
└── cmd/logstatsd/          ← you implement serve(); run/main/newLogger provided
```

## Config precedence + env vars

| Layer | Source | Example |
|---|---|---|
| defaults | baked in | `:8080`, `:9090`, `info` |
| file | `-config` / `LOGSTATS_CONFIG` (JSON) | `{"log_level":"warn"}` |
| env | `LOGSTATS_HTTP_ADDR`, `LOGSTATS_GRPC_ADDR`, `LOGSTATS_LOG_LEVEL`, `LOGSTATS_AUTH_TOKEN` | `LOGSTATS_LOG_LEVEL=debug` |
| flags | `-http-addr`, `-grpc-addr`, `-log-level`, `-config` | `-log-level=error` |

Each layer overrides the ones before it; flags win. **`AuthToken` is env-only** (no flag — flags leak in `ps`) and is **redacted** in logs.

---

## Concept 1 — Config precedence

The same binary runs everywhere with different settings, so config is layered: **defaults < file < env < flags**. Defaults let it run with zero config; a file carries per-environment settings; env is how containers inject config (12-factor); flags are the operator's highest-priority, most-explicit override.

```go
cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // defaults
// overlay file (if -config/LOGSTATS_CONFIG), then env, then explicitly-set flags
```

`Load(args, getenv)` takes an injected `getenv func(string) string` so tests pass a fake environment.

### Common mistake

No precedence (only flags, or only env — can't layer), or **precedence inversion** where a config file silently overrides an explicit `-flag` an operator typed to debug a live issue. Flags are direct operator intent; they win.

---

## Concept 2 — The `flag` package + env

Use a `flag.FlagSet` (not the package-global `flag`) for self-contained, testable loading. The crux of precedence is `fs.Visit`, which iterates **only explicitly-set flags**:

```go
fs := flag.NewFlagSet("logstatsd", flag.ContinueOnError)
httpAddr := fs.String("http-addr", "", "HTTP listen address")
fs.Parse(args)
// ... apply file + env first ...
fs.Visit(func(f *flag.Flag) {        // only flags the user actually typed
    if f.Name == "http-addr" { cfg.HTTPAddr = *httpAddr }
})
```

A flag the user typed overrides env/file; a flag left at its default doesn't touch them — which is exactly why flags sit at the top of the stack.

### Common mistake

Using the package-global `flag` (reads `os.Args`, panics on re-parse, leaks state between tests). And **treating a defaulted flag as set** — reading `*httpAddr` unconditionally clobbers the env/file value with an empty default, inverting precedence. Use `fs.Visit`.

---

## Concept 3 — Config files (JSON, stdlib)

A richer config wants a file; `encoding/json` into your struct is enough — no third-party lib.

```go
if path != "" {
    data, err := os.ReadFile(path)
    if err != nil { return Config{}, fmt.Errorf("config: read %s: %w", path, err) }
    if err := json.Unmarshal(data, &cfg); err != nil {
        return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
    }
}
```

Unmarshalling into the *already-defaulted* `cfg` overlays only the fields present in the file; absent fields keep their defaults. The file path is itself bootstrapped from a flag/env.

### Common mistake

Applying the file *after* flags (so the file overrides explicit flags — inversion again). Apply file early: defaults → **file** → env → flags. And wrap file errors with the path (`%w` + filename), or "no such file" with no name is miserable to debug.

---

## Concept 4 — Secrets hygiene

Secrets (tokens, passwords) need the precedence machinery plus care:

1. **From env/file, never a flag** — flags show in `ps aux` and shell history. So `AuthToken` reads `LOGSTATS_AUTH_TOKEN` only; there is no `-auth-token` flag.
2. **Never logged** — the danger is logging the whole config struct for debugging. `slog.LogValuer` controls how `Config` appears in logs:

```go
func (c Config) LogValue() slog.Value {
    token := ""
    if c.AuthToken != "" { token = "***" }
    return slog.GroupValue(
        slog.String("http_addr", c.HTTPAddr),
        slog.String("grpc_addr", c.GRPCAddr),
        slog.String("log_level", c.LogLevel),
        slog.String("auth_token", token), // never the real value
    )
}
```

Now `logger.Info("starting", "config", cfg)` prints `auth_token:"***"` — operators see a token *is* set without seeing its value.

### Common mistake

Logging the whole config struct without a `LogValuer` (`%+v` dumps the token). Or putting secrets in flags / committed config files — both end up somewhere searchable. The value lives only in the environment, and appears only redacted.

---

## Concept 5 — Graceful shutdown

On SIGTERM (deploy, scale-down, node drain) the orchestrator waits a grace period then SIGKILLs. Exit immediately and you drop in-flight requests — 500s mid-deploy. Graceful = stop accepting, finish in-flight (bounded), exit zero.

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
// in serve(), on <-ctx.Done():
shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
httpSrv.Shutdown(shutCtx)   // drain HTTP: stop accepting, wait for active handlers
grpcSrv.GracefulStop()      // drain gRPC: wait for active RPCs/streams
```

`logstatsd`'s `serve` runs both servers in goroutines (fatal errors on a buffered `errCh`) and `select`s on `ctx.Done()` vs `errCh`: on signal, drain both with a 10s budget and return `nil`; if one server crashes first, tear down the other and return its error.

### Common mistake

`os.Exit` on signal (drops in-flight work instantly — the deploy-time 500 storm), or no shutdown timeout (`Shutdown(context.Background())` hangs forever on one stuck client, so you get SIGKILLed having drained nothing). Always bound the drain.

---

## Exercise: warm-up — `config`

Implement `Load(args, getenv) (Config, error)` in `exercises/warmup/config/config.go`: defaults < JSON file < env < explicitly-set flags (use `fs.Visit`), validate the log level. `Config` + `LogValue` redaction are provided.

**Time:** 15-20 minutes.

## Exercise: main — `cmd/logstatsd`

Implement `serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis)`: run both servers concurrently; on `ctx.Done()` drain both (`http.Shutdown` bounded + `grpc.GracefulStop`); if one errors first, tear down the other and return the error. `run`/`main`/`newLogger` are provided (they wire config → servers → listeners and log the redacted config).

**Time:** 30-45 minutes.

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18 — the shared Store + concurrent servers
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/26-config/exercises
go test ./...

# Reference solution
go test -v ./lessons/26-config/solutions/...
go test -race ./lessons/26-config/solutions/...
make test-race

# Run the daemon (config via flag/env/file) and shut it down with SIGTERM
go build -o /tmp/logstatsd ./lessons/26-config/solutions/cmd/logstatsd
LOGSTATS_LOG_LEVEL=debug /tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
# {"accepted":2,"parsed":1,"failed":1}
curl -s 127.0.0.1:8080/stats; echo
# {"counts":{"INFO":1},"total":1}
kill -TERM %1   # watch it drain cleanly (startup log shows auth_token:"***" if set)
rm -f /tmp/logstatsd
```

## Going further

- **A `-version` flag** stamped via `-ldflags -X` (a preview of L27); print it and exit.
- **SIGHUP config reload** — re-run `Load` on SIGHUP and swap the log level live (without dropping connections).
- **Readiness gating during drain** — make `/healthz` (or a new `/readyz`) return 503 the moment shutdown begins, so a load balancer stops routing before the drain finishes.
- **`--dry-run`** — load config, print the resolved (redacted) `Config`, and exit 0; handy for verifying precedence in CI.
- **Per-listener drain budgets** — give HTTP and gRPC independent timeouts and run the two drains concurrently.

---

> Phase 4 continues. Next: Lesson 27 — **Build, release & containerization**. We ship `logstatsd`: `-ldflags` version stamping, static `CGO_ENABLED=0` cross-compiled binaries, a multi-stage Dockerfile on distroless, and image hygiene. A configurable, gracefully-shutting-down daemon is exactly what an orchestrator wants to run.
