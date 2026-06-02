# Plan CC — Lesson 26 (Configuration, secrets & graceful shutdown) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.
>
> **Commit policy:** NO `Co-Authored-By` trailer or "Generated with" line in commit messages or PR bodies. Subject + body only.

**Goal:** Author lesson 26 — the fourth Phase 4 lesson. Turn the `logstats` service into a real daemon: a single `logstatsd` process serving HTTP **and** gRPC over one shared `logstats.Store`, configured by flags > env > file precedence, keeping secrets out of logs, and draining both faces gracefully on SIGINT/SIGTERM.

**Architecture:** Same per-lesson pattern, six tasks. Carries L25's `logparse` + `logstats` + `grpcsrv` + proto/generated code verbatim (keeps grpc deps live). Extracts L25's HTTP server into a reusable `internal/httpsrv` package. New: `warmup/config` (the precedence loader + secret redaction) and `cmd/logstatsd` (the unified daemon with coordinated graceful shutdown). Retires the three separate L25 cmd binaries. Five-concept slide deck.

**Tech Stack:** Go 1.23 stdlib (`flag`, `encoding/json`, `log/slog`, `os`, `os/signal`, `syscall`, `net`, `net/http`, `context`, `time`) + the carried `google.golang.org/grpc` v1.73.0 / `protobuf` v1.36.11 (no NEW deps).

---

## Scope

After Plan CC: lesson 26 complete; `make test` + `make test-race` green; `logstatsd` serves both faces, loads config with correct precedence, redacts secrets in logs, and drains cleanly on signal; `26-config` in the index; go directive stays `go 1.23`.

### Design decisions (2 user-approved + plan-recommended)

**User-approved via brainstorming:**

1. **Unified `logstatsd`** — one process serving HTTP + gRPC over a shared `Store`, coordinated run loop, draining both on SIGINT/SIGTERM. Retires the three separate L25 binaries.
2. **Five slide concepts:** config precedence · `flag` package + env · config files (JSON, stdlib) · secrets hygiene · graceful shutdown.

**Plan-recommended:**

3. **Carry the gRPC face forward** (proto + generated + grpcsrv) so `go.mod` keeps importing grpc/protobuf — otherwise `go mod tidy` drops the deps and L27+ re-triggers the version-pinning dance.
4. **Extract `internal/httpsrv`** from L25's `cmd/logstats-server` (the router/handlers/middleware) so the daemon can mount the HTTP face. Mechanical refactor of already-verified code; parallels `grpcsrv`.
5. **`warmup/config` imported by the daemon** (precedent: L24's `warmup/retry` imported by `internal/shipper`). `Load(args, getenv)` takes injected `getenv` for testability.
6. **Precedence: defaults < JSON file < env < explicitly-set flags**, via `flag.FlagSet` + `fs.Visit` (only set flags override). Secret `AuthToken` is **env-only** (never a flag — flags leak in `ps`) and **redacted** via `slog.LogValuer`.
7. **`run`/`serve` split** (the L23 pattern, extended to two listeners): `run` wires config→servers→listeners (provided); `serve` is the coordinated-drain exercise.

### Verified facts (prototyped before writing this plan)

- `config.Load` precedence (defaults/file/env/flag layering, flag-beats-env-beats-file), invalid-level error, and `slog.LogValuer` redaction — all pass.
- `serve` coordinated drain: HTTP + gRPC on two `127.0.0.1:0` listeners, HTTP request succeeds, gRPC dial succeeds, `ctx` cancel → both drain → `serve` returns `nil`, HTTP closed after. Race-clean over 8 runs.
- `grpc.Server.Serve` returns `nil` after `GracefulStop`; `http.Server.Serve` returns `http.ErrServerClosed` after `Shutdown` (swallowed).

---

## File structure

```
lessons/26-config/{exercises,solutions}/
├── warmup/config/{config.go, config_test.go}        (Task 3 — Load + LogValue SKELETON)
├── proto/logstats.proto + logstatspb/*.pb.go         (Task 1 — carried, regenerated, committed)
├── internal/
│   ├── logparse/                                     (Task 1 — carried verbatim)
│   ├── logstats/                                     (Task 1 — carried verbatim)
│   ├── grpcsrv/                                      (Task 1 — carried verbatim)
│   └── httpsrv/{httpsrv.go, httpsrv_test.go}          (Task 2 — extracted from L25 HTTP server)
└── cmd/logstatsd/{main.go, main_test.go}             (Task 4 — run provided; serve SKELETON)
```

---

## Task 1: Scaffold + carry forward + regenerate proto (controller-direct)

- [ ] **Step 1:** Scaffold + remove flat stubs:
```bash
make new-lesson NAME=26-config
rm lessons/26-config/exercises/main.go lessons/26-config/exercises/main_test.go \
   lessons/26-config/exercises/warmup.go lessons/26-config/exercises/warmup_test.go \
   lessons/26-config/solutions/main.go lessons/26-config/solutions/main_test.go \
   lessons/26-config/solutions/warmup.go lessons/26-config/solutions/warmup_test.go
```

- [ ] **Step 2:** Carry forward logparse + logstats + grpcsrv + proto (both trees), rewrite paths + header:
```bash
SRC=lessons/25-grpc
DST=lessons/26-config
for side in exercises solutions; do
  mkdir -p "$DST/$side/internal" "$DST/$side/proto"
  cp -R "$SRC/$side/internal/logparse"  "$DST/$side/internal/logparse"
  cp -R "$SRC/$side/internal/logstats"  "$DST/$side/internal/logstats"
  cp -R "$SRC/$side/internal/grpcsrv"   "$DST/$side/internal/grpcsrv"
  cp "$SRC/$side/proto/logstats.proto"  "$DST/$side/proto/logstats.proto"
done
grep -rl '25-grpc' "$DST" | while read -r f; do sed -i '' 's#lessons/25-grpc#lessons/26-config#g' "$f"; done
grep -rl 'lesson 25' "$DST" | while read -r f; do sed -i '' 's/lesson 25/lesson 26/g' "$f"; done
# also fix the go_package in the carried .proto (sed above already did the path)
```

- [ ] **Step 3:** Update the `Makefile` `proto` target to ALSO generate the L26 protos (append four lines for `lessons/26-config/{exercises,solutions}/proto/logstats.proto` — note L26 has no greeter). Then regenerate:
```bash
make proto
go mod tidy   # grpc/protobuf still imported by grpcsrv → kept
grep '^go ' go.mod   # MUST stay go 1.23.0
```

- [ ] **Step 4:** Verify the carried baseline (logparse/logstats/grpcsrv tests pass; generated code builds):
```bash
go build ./lessons/26-config/...
go test ./lessons/26-config/... 2>&1 | tail -12
go vet ./lessons/26-config/... && golangci-lint run ./lessons/26-config/... && gofmt -l lessons/26-config/
```
Expected: build clean; carried tests pass; vet/lint/fmt clean.

- [ ] **Step 5:** Commit:
```bash
git add lessons/26-config Makefile go.mod go.sum
git commit -m "chore(lesson-26): scaffold + carry forward L25 logparse/logstats/grpcsrv/proto"
```

> Note: the carried `grpcsrv` keeps grpc/protobuf imported, so `go mod tidy` retains the deps and the directive stays `go 1.23`. There is no `warmup/greeter` in L26 (the greeter was L25-only) — do NOT carry it; remove the L26 greeter proto lines if the scaffolder/sed created any.

---

## Task 2: Extract `internal/httpsrv` from the carried HTTP server

L25's `cmd/logstats-server/main.go` was carried in Task 1. Extract its router/handlers/middleware into `internal/httpsrv` so the daemon can mount the HTTP face, then DELETE the carried `cmd/logstats-server` (replaced by `logstatsd`).

**Files (4 total) + 1 deletion per tree.**

- [ ] **Step 1:** Create `lessons/26-config/{exercises,solutions}/internal/httpsrv/httpsrv.go` by moving, from the carried `cmd/logstats-server/main.go`, everything EXCEPT `run`/`serve`/`main`: the `ingestRequest`/`ingestResponse`/`statsResponse` types, `newRouter` (rename → exported **`Router`**), `ingestHandler`, `statsHandler`, `healthHandler`, `writeJSON`, `statusRecorder`, `withRequestLog`, `withRecovery`. Package `httpsrv`. Imports `logparse` + `logstats` (per-tree paths). The exported entry point:

```go
// Package httpsrv builds the logstats HTTP router — the service's HTTP
// face, extracted so the unified daemon can mount it alongside gRPC.
package httpsrv

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/ristkari-dev/go-training/lessons/26-config/<tree>/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/26-config/<tree>/internal/logstats"
)

const maxIngestBytes = 1 << 20

// Router returns the HTTP handler for the logstats service (POST /ingest,
// GET /stats, GET /healthz), wrapped in logging + recovery middleware.
func Router(store *logstats.Store, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingestHandler(store))
	mux.HandleFunc("GET /stats", statsHandler(store))
	mux.HandleFunc("GET /healthz", healthHandler)
	return withRequestLog(withRecovery(mux, logger), logger)
}

// ... (ingestRequest/ingestResponse/statsResponse, ingestHandler,
// statsHandler, healthHandler, writeJSON, statusRecorder, withRequestLog,
// withRecovery — verbatim from the carried cmd/logstats-server/main.go) ...
```

(In the EXERCISES tree, `httpsrv` is provided fully working — it's carried code, not an exercise.)

- [ ] **Step 2:** Create `internal/httpsrv/httpsrv_test.go` (both trees, real tests) by moving the handler tests from the carried `cmd/logstats-server/main_test.go` that don't depend on `run`/`serve`: `TestIngestThenStats`, `TestIngestBadJSON`, `TestHealthz`, `TestMethodNotAllowed`, `TestConcurrentIngestRace` — change `testRouter()` to call `Router(logstats.NewStore(), discardLogger)`. Drop `TestServeLifecycle` (that logic moves to the daemon).

- [ ] **Step 3:** Delete the carried HTTP server cmd (both trees):
```bash
rm -r lessons/26-config/exercises/cmd/logstats-server lessons/26-config/solutions/cmd/logstats-server
```

- [ ] **Step 4:** Verify + commit:
```bash
gofmt -l lessons/26-config/
go test ./lessons/26-config/.../internal/httpsrv/... 2>&1 | tail -8
go test -race ./lessons/26-config/solutions/internal/httpsrv/... 2>&1 | tail -3
make test && go vet ./lessons/26-config/... && golangci-lint run ./lessons/26-config/...

git add lessons/26-config
git commit -m "refactor(lesson-26): extract internal/httpsrv from the HTTP server cmd"
```

Expected: httpsrv tests pass in both trees; -race clean.

---

## Task 3: Warm-up — `config`

`Load(args, getenv) (Config, error)` with precedence + `slog.LogValuer` redaction. (VERIFIED code.)

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/26-config/{exercises,solutions}/warmup/config`

- [ ] **Step 2:** `exercises/warmup/config/config.go` (SKELETON — `Config` type + `LogValue` + `validLevel` provided; `Load` panics):

```go
// Package config loads logstatsd configuration with precedence:
// defaults < JSON file < environment < explicitly-set flags.
package config

import (
	"flag"
	"fmt"
	"log/slog"
)

// Config holds the logstatsd runtime configuration.
type Config struct {
	HTTPAddr  string `json:"http_addr"`
	GRPCAddr  string `json:"grpc_addr"`
	LogLevel  string `json:"log_level"`
	AuthToken string `json:"-"` // secret: env-only, never logged
}

// LogValue implements slog.LogValuer so logging a Config never leaks the
// secret token — it shows "***" when set, "" when empty.
func (c Config) LogValue() slog.Value {
	token := ""
	if c.AuthToken != "" {
		token = "***"
	}
	return slog.GroupValue(
		slog.String("http_addr", c.HTTPAddr),
		slog.String("grpc_addr", c.GRPCAddr),
		slog.String("log_level", c.LogLevel),
		slog.String("auth_token", token),
	)
}

func validLevel(s string) bool {
	switch s {
	case "debug", "info", "warn", "error":
		return true
	}
	return false
}

// Load resolves configuration with precedence: defaults < JSON file <
// env < explicitly-set flags. getenv is injected for testability.
//
// Hint:
//   1. cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // defaults
//   2. fs := flag.NewFlagSet(...); define -config -http-addr -grpc-addr -log-level; fs.Parse(args)
//   3. file layer: path from -config or getenv("LOGSTATS_CONFIG"); os.ReadFile + json.Unmarshal into &cfg
//   4. env layer: LOGSTATS_HTTP_ADDR / _GRPC_ADDR / _LOG_LEVEL / _AUTH_TOKEN (non-empty overrides)
//   5. flag layer: fs.Visit(...) — only EXPLICITLY-SET flags override (not defaulted ones)
//   6. validate cfg.LogLevel; return error if invalid
func Load(args []string, getenv func(string) string) (Config, error) {
	_ = flag.NewFlagSet
	_ = fmt.Errorf
	_ = validLevel
	panic("TODO: defaults < file < env < explicitly-set flags; validate level")
}
```

- [ ] **Step 3:** `exercises/warmup/config/config_test.go` (SKELETON):

```go
package config

import "testing"

// TestLoad is a SKELETON. Cover precedence: defaults; env over default;
// flag over env; file over default; env over file; flag over file; an
// invalid level errors. Plus a redaction test: logging a Config with an
// AuthToken must not leak it (shows "***").
func TestLoad(t *testing.T) {
	// TODO: env := func(m map[string]string) func(string) string { ... }
	//       cfg, err := Load(nil, env(nil)); assert defaults
	_ = Load
}
```

- [ ] **Step 4:** `solutions/warmup/config/config.go` — same as Step 2 but with `Load` implemented (add imports `encoding/json`, `os`):

```go
// Package config loads logstatsd configuration with precedence:
// defaults < JSON file < environment < explicitly-set flags.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	HTTPAddr  string `json:"http_addr"`
	GRPCAddr  string `json:"grpc_addr"`
	LogLevel  string `json:"log_level"`
	AuthToken string `json:"-"`
}

func (c Config) LogValue() slog.Value {
	token := ""
	if c.AuthToken != "" {
		token = "***"
	}
	return slog.GroupValue(
		slog.String("http_addr", c.HTTPAddr),
		slog.String("grpc_addr", c.GRPCAddr),
		slog.String("log_level", c.LogLevel),
		slog.String("auth_token", token),
	)
}

func validLevel(s string) bool {
	switch s {
	case "debug", "info", "warn", "error":
		return true
	}
	return false
}

func Load(args []string, getenv func(string) string) (Config, error) {
	cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info"} // defaults

	fs := flag.NewFlagSet("logstatsd", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON config file")
	httpAddr := fs.String("http-addr", "", "HTTP listen address")
	grpcAddr := fs.String("grpc-addr", "", "gRPC listen address")
	logLevel := fs.String("log-level", "", "log level (debug|info|warn|error)")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	// Layer 1: JSON config file (from -config flag or LOGSTATS_CONFIG env).
	path := *configPath
	if path == "" {
		path = getenv("LOGSTATS_CONFIG")
	}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("config: read %s: %w", path, err)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
		}
	}

	// Layer 2: environment overrides.
	if v := getenv("LOGSTATS_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := getenv("LOGSTATS_GRPC_ADDR"); v != "" {
		cfg.GRPCAddr = v
	}
	if v := getenv("LOGSTATS_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := getenv("LOGSTATS_AUTH_TOKEN"); v != "" {
		cfg.AuthToken = v
	}

	// Layer 3: explicitly-set flags win (fs.Visit skips defaulted flags).
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "http-addr":
			cfg.HTTPAddr = *httpAddr
		case "grpc-addr":
			cfg.GRPCAddr = *grpcAddr
		case "log-level":
			cfg.LogLevel = *logLevel
		}
	})

	if !validLevel(cfg.LogLevel) {
		return Config{}, fmt.Errorf("config: invalid log level %q", cfg.LogLevel)
	}
	return cfg, nil
}
```

- [ ] **Step 5:** `solutions/warmup/config/config_test.go` (full — VERIFIED):

```go
package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadPrecedence(t *testing.T) {
	cfg, err := Load(nil, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.GRPCAddr != ":9090" || cfg.LogLevel != "info" {
		t.Errorf("defaults wrong: %+v", cfg)
	}

	cfg, _ = Load(nil, env(map[string]string{"LOGSTATS_HTTP_ADDR": ":7000", "LOGSTATS_LOG_LEVEL": "debug"}))
	if cfg.HTTPAddr != ":7000" || cfg.LogLevel != "debug" {
		t.Errorf("env override failed: %+v", cfg)
	}

	cfg, _ = Load([]string{"-http-addr=:6000"}, env(map[string]string{"LOGSTATS_HTTP_ADDR": ":7000"}))
	if cfg.HTTPAddr != ":6000" {
		t.Errorf("flag should beat env: %+v", cfg)
	}
}

func TestLoadFileLayer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(path, []byte(`{"http_addr":":5000","grpc_addr":":5001","log_level":"warn"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load([]string{"-config=" + path}, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":5000" || cfg.GRPCAddr != ":5001" || cfg.LogLevel != "warn" {
		t.Errorf("file layer failed: %+v", cfg)
	}

	cfg, _ = Load([]string{"-config=" + path, "-log-level=error"},
		env(map[string]string{"LOGSTATS_GRPC_ADDR": ":5999"}))
	if cfg.HTTPAddr != ":5000" {
		t.Errorf("file value lost: %+v", cfg)
	}
	if cfg.GRPCAddr != ":5999" {
		t.Errorf("env should beat file: %+v", cfg)
	}
	if cfg.LogLevel != "error" {
		t.Errorf("flag should beat file: %+v", cfg)
	}
}

func TestLoadInvalidLevel(t *testing.T) {
	if _, err := Load([]string{"-log-level=verbose"}, env(nil)); err == nil {
		t.Error("expected error for invalid level")
	}
}

func TestConfigRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := Config{HTTPAddr: ":8080", GRPCAddr: ":9090", LogLevel: "info", AuthToken: "super-secret-xyz"}
	logger.Info("config loaded", "config", cfg)

	out := buf.String()
	if strings.Contains(out, "super-secret-xyz") {
		t.Errorf("secret leaked in log: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("redaction marker missing: %s", out)
	}
}
```

- [ ] **Step 6:** Verify + commit:
```bash
gofmt -l lessons/26-config/
go test ./lessons/26-config/exercises/warmup/config/... 2>&1 | tail -5
go test -v ./lessons/26-config/solutions/warmup/config/... 2>&1 | tail -20
make test && go vet ./lessons/26-config/... && golangci-lint run ./lessons/26-config/...

git add lessons/26-config/exercises/warmup lessons/26-config/solutions/warmup
git commit -m "feat(lesson-26): warmup — config (flags>env>file precedence + slog redaction)"
```

Expected: exercises vacuous-pass; solutions 4 tests PASS.

---

## Task 4: `cmd/logstatsd` — the unified daemon

`run` (provided) wires config → HTTP + gRPC servers → listeners; `serve` (the EXERCISE) runs both and drains both on signal.

**Files (4 total).** Per-tree import paths.

- [ ] **Step 1:** `mkdir -p lessons/26-config/{exercises,solutions}/cmd/logstatsd`

- [ ] **Step 2:** `exercises/cmd/logstatsd/main.go` (SKELETON — everything provided EXCEPT `serve`, which panics). Exercises import paths:

```go
// Package main is the logstatsd daemon: it serves the logstats HTTP and
// gRPC faces from one process over a shared store, configured by
// flags>env>file, and drains both gracefully on SIGINT/SIGTERM.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/26-config/exercises/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/exercises/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/exercises/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/26-config/exercises/proto/logstatspb"
	"github.com/ristkari-dev/go-training/lessons/26-config/exercises/warmup/config"
)

const drainTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load(os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newLogger(level string, w io.Writer) *slog.Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l}))
}

// run wires config → servers → listeners, logs the (redacted) config,
// and hands off to serve. Provided.
func run(ctx context.Context, cfg config.Config, stdout io.Writer) error {
	logger := newLogger(cfg.LogLevel, stdout)
	logger.Info("starting logstatsd", "config", cfg) // redacted via Config.LogValue

	store := logstats.NewStore()

	httpLis, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("http listen: %w", err)
	}
	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	httpSrv := &http.Server{Handler: httpsrv.Router(store, logger)}
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcsrv.LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(grpcsrv.LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(grpcSrv, grpcsrv.New(store))

	logger.Info("listening", "http", httpLis.Addr().String(), "grpc", grpcLis.Addr().String())
	return serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis)
}

// serve runs both servers concurrently and drains BOTH when ctx is
// cancelled. If either fails first, stop the other and return the error.
// IMPLEMENT THIS.
//
// Hint:
//   errCh := make(chan error, 2)
//   go func(){ if err := httpSrv.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) { errCh <- err } }()
//   go func(){ if err := grpcSrv.Serve(grpcLis); err != nil { errCh <- err } }()
//   select {
//   case <-ctx.Done():           // graceful drain
//       shutCtx, cancel := context.WithTimeout(context.Background(), drainTimeout); defer cancel()
//       httpSrv.Shutdown(shutCtx); grpcSrv.GracefulStop(); return nil
//   case err := <-errCh:         // one failed → tear down the other
//       ... httpSrv.Shutdown(...); grpcSrv.GracefulStop(); return err
//   }
func serve(ctx context.Context, httpSrv *http.Server, httpLis net.Listener, grpcSrv *grpc.Server, grpcLis net.Listener) error {
	_ = drainTimeout
	panic("TODO: run both servers; on ctx.Done drain both (http.Shutdown + grpc.GracefulStop); on server error tear down the other")
}
```

- [ ] **Step 3:** `exercises/cmd/logstatsd/main_test.go` (SKELETON — does not call run/serve):

```go
package main

import "testing"

// TestServe is a SKELETON. Stand up two 127.0.0.1:0 listeners + an
// http.Server (httpsrv.Router) + a grpc.Server (grpcsrv registered),
// run serve in a goroutine, hit /healthz over HTTP and GetStats over
// gRPC, cancel the ctx, and assert serve returns nil + both stopped.
func TestServe(t *testing.T) {
	// TODO — see the solution.
}
```

- [ ] **Step 4:** `solutions/cmd/logstatsd/main.go` — identical to Step 2 but `serve` implemented + `errors` import + `.../solutions/...` paths:

```go
// (same header + imports as exercises, plus "errors", solutions paths)

func serve(ctx context.Context, httpSrv *http.Server, httpLis net.Listener, grpcSrv *grpc.Server, grpcLis net.Listener) error {
	errCh := make(chan error, 2)
	go func() {
		if err := httpSrv.Serve(httpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go func() {
		if err := grpcSrv.Serve(grpcLis); err != nil {
			errCh <- err
		}
	}()

	drain := func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
		defer cancel()
		_ = httpSrv.Shutdown(shutCtx)
		grpcSrv.GracefulStop()
	}

	select {
	case <-ctx.Done():
		drain()
		return nil
	case err := <-errCh:
		drain()
		return err
	}
}
```

(The rest of `solutions/cmd/logstatsd/main.go` — `main`, `newLogger`, `run` — is identical to the exercises file with `solutions` import paths and the added `"errors"` import.)

- [ ] **Step 5:** `solutions/cmd/logstatsd/main_test.go` (real coordinated-drain integration test):

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/26-config/solutions/proto/logstatspb"
)

func TestServe(t *testing.T) {
	httpLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	store := logstats.NewStore()
	httpSrv := &http.Server{Handler: httpsrv.Router(store, logger)}
	grpcSrv := grpc.NewServer()
	pb.RegisterLogStatsServer(grpcSrv, grpcsrv.New(store))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis) }()

	// HTTP face works.
	hurl := fmt.Sprintf("http://%s/healthz", httpLis.Addr().String())
	ready := false
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); {
		if resp, err := http.Get(hurl); err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("HTTP never became ready")
	}

	// gRPC face works.
	conn, err := grpc.NewClient(grpcLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()
	if _, err := pb.NewLogStatsClient(conn).GetStats(context.Background(), &pb.StatsRequest{}); err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	// Cancel → both drain → serve returns nil.
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("serve returned %v, want nil", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("serve did not return after cancel")
	}
	if _, err := http.Get(hurl); err == nil {
		t.Error("HTTP still serving after drain")
	}
}
```

- [ ] **Step 6:** Verify + commit:
```bash
gofmt -l lessons/26-config/
go test ./lessons/26-config/exercises/cmd/logstatsd/... 2>&1 | tail -5
go test -v ./lessons/26-config/solutions/cmd/logstatsd/... 2>&1 | tail -10
go test -race ./lessons/26-config/solutions/cmd/logstatsd/... 2>&1 | tail -3
make test && make test-race && go vet ./lessons/26-config/... && golangci-lint run ./lessons/26-config/...

# Manual end-to-end (config precedence + both faces + graceful shutdown)
go build -o /tmp/logstatsd ./lessons/26-config/solutions/cmd/logstatsd
LOGSTATS_LOG_LEVEL=debug /tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s 127.0.0.1:8080/stats; echo
kill -TERM %1; wait %1 2>/dev/null; echo "exit: $?"
rm -f /tmp/logstatsd

git add lessons/26-config/exercises/cmd lessons/26-config/solutions/cmd
git commit -m "feat(lesson-26): cmd/logstatsd (unified HTTP+gRPC daemon, config, graceful drain)"
```

Expected: exercises vacuous-pass; solutions TestServe PASS; -race clean; smoke shows ingest/stats working then a clean SIGTERM shutdown.

---

## Task 5: Slide deck — 5 concepts (controller inline)

**File:** `lessons/26-config/slides/slides.md`. Concepts:

1. **Config precedence** — defaults < file < env < flags; the 12-factor "config in the environment"; why this order (flags = explicit operator intent, highest). *Mistake:* no precedence / hardcoded config.
2. **`flag` package + env** — `FlagSet`, `fs.Visit` vs `VisitAll` (knowing what was explicitly set is the key), injected `getenv` for tests. *Mistake:* package-global `flag` (untestable); treating a defaulted flag as "set" (clobbers env/file).
3. **Config files (JSON, stdlib)** — `encoding/json` into the struct; layer it under env+flags; no Viper needed. *Mistake:* a file silently overriding an explicit flag (precedence inversion).
4. **Secrets hygiene** — secrets from env/file, never flags (`ps` leak); never log them; `slog.LogValuer` redaction; don't commit them. *Mistake:* logging the whole config struct → token in the logs.
5. **Graceful shutdown** — `signal.NotifyContext(SIGINT, SIGTERM)`; coordinated drain of HTTP (`Shutdown`) + gRPC (`GracefulStop`) with a timeout; finish in-flight, reject new. *Mistake:* `os.Exit` on signal (drops in-flight work); no drain timeout (hangs on a stuck conn).

- [ ] Author the deck (title-slide-grid, "What we'll cover", "The story so far" — from binaries to a daemon, 5 concepts, Practice, Closing thought, What we learned, Up next → L27 containerization). Then:
```bash
make slides-build && grep -q "26-config" dist/index.html && echo "✓ index" && rm -rf dist
git add lessons/26-config/slides/
git commit -m "feat(lesson-26): slides — config, secrets & graceful shutdown (5 concepts)"
```

---

## Task 6: README + verify + final review + PR (controller)

- [ ] **Step 1:** Confirm `tools/build-index/main.go` lists L26 as `{Number:"26", Slug:"config", Title:"Config & shutdown", Blurb:"flags · env · signals", Phase:4}` — slug matches; no change (`grep '"26"' tools/build-index/main.go`).

- [ ] **Step 2:** Write `lessons/26-config/README.md` (~300 lines): "What's different from L25" (three binaries → one configurable daemon; HTTP face extracted to `internal/httpsrv`); the config precedence table + env var names + the secret-token rule; the graceful-shutdown drain sequence; how to run (config via flag/env/file; SIGTERM); "going further" (a `-version` flag (preview of L27); SIGHUP config reload; readiness vs liveness gating during drain; a `--dry-run` that prints resolved config; per-listener drain timeouts).

- [ ] **Step 3:** Full sweep:
```bash
make test && make test-race
go test -v ./lessons/26-config/solutions/...
go test -race ./lessons/26-config/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/26-config/
grep '^go ' go.mod   # still go 1.23.0
make slides-build && grep -q "26-config" dist/index.html && rm -rf dist
```

- [ ] **Step 4:** Commit README:
```bash
git add lessons/26-config/README.md
git commit -m "docs(lesson-26): README — config, secrets & graceful shutdown self-study"
```

- [ ] **Step 5:** Dispatch `feature-dev:code-reviewer` over `git diff main...HEAD`. Focus: config precedence correctness (fs.Visit semantics; file<env<flag layering; the `json:"-"` on AuthToken so a config file can't set the secret — confirm that's intended); redaction completeness (no path logs the raw struct); the serve coordinated-drain (no goroutine leak; errCh buffering; drain on both ctx-cancel and server-error paths; grpc.Serve returning nil after GracefulStop); httpsrv extraction parity with L25; test determinism (ephemeral ports, 15s ceiling); deps unchanged (go 1.23); exercises skeletons vacuous. Apply fixes.

- [ ] **Step 6:** Push + open PR (no co-author trailer); watch CI green.

---

## Verification (after Task 6)

```bash
make test && make test-race
go test ./lessons/26-config/exercises/...                 # vacuous-pass
go test -v ./lessons/26-config/solutions/...              # config + httpsrv + grpcsrv + logstatsd + carried
go test -race ./lessons/26-config/...
go vet ./... && golangci-lint run ./...
gofmt -l lessons/26-config/ && grep '^go ' go.mod         # go 1.23.0
make slides-build && grep -q "26-config" dist/index.html && rm -rf dist

# Manual: config precedence + both faces + SIGTERM drain
go build -o /tmp/logstatsd ./lessons/26-config/solutions/cmd/logstatsd
LOGSTATS_LOG_LEVEL=debug /tmp/logstatsd -http-addr=127.0.0.1:8080 -grpc-addr=127.0.0.1:9090 &
sleep 0.5
curl -s -XPOST 127.0.0.1:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s 127.0.0.1:8080/stats; echo
kill -TERM %1; wait %1 2>/dev/null
rm -f /tmp/logstatsd
```

## Critical file paths

To create: `lessons/26-config/` tree (README, slides, warmup/config, internal/{logparse,logstats,grpcsrv,httpsrv}, proto+logstatspb, cmd/logstatsd, mirrored).
To modify: `Makefile` (proto target += L26 protos), `go.mod`/`go.sum` (tidy; no new direct deps).
To delete (after carry): the carried `cmd/logstats-server` (replaced by `logstatsd`).
To reference: `lessons/25-grpc/...` (carry-forward source), `tools/build-index/main.go` (slug `config` already correct).

## Execution after approval

1. (Branch `feature/plan-cc-lesson-26-config` already created off main.)
2. Commit this plan doc.
3. Execute: Task 1 controller-direct (scaffold + carry + regenerate — environment-sensitive); subagents for Tasks 2-4 (httpsrv extraction, config, logstatsd); controller inline for Tasks 5-6 (slides, README); controller runs all verification.
4. Final code review subagent. Push + PR (no co-author trailer); watch CI green.
