# Lesson 23: HTTP servers

## What you'll learn

By the end of this lesson you can:

- Build an HTTP server with **`net/http`** and Go 1.22 **`ServeMux`** method+pattern routing (no framework).
- Write **handlers** that decode/encode JSON, set status codes, and cap request bodies.
- Factor cross-cutting concerns into **middleware** (`func(http.Handler) http.Handler`) — logging + panic recovery.
- Emit **structured logs** with `log/slog`.
- Design a service around **concurrency-safe shared state** and make it testable with `net/http/httptest`.

This is the start of **Phase 4 — Production & Distributed**. The L16-22 log aggregator becomes the **`logstats` service**.

## What's different — Phase 4 begins

Phases 1-3 built the language and ran the aggregator as a batch tool that walks a directory once. **L23 turns it into a long-lived service.** Instead of `Walk(dir)`, clients `POST /ingest` log lines; the server parses them with the carried-forward `logparse` and accumulates per-level counts in shared memory that any client can read via `GET /stats`. Over Phase 4 this one service gains a resilient client (L24), a gRPC API (L25), config + graceful shutdown (L26), a container (L27), observability (L28), and idempotent distributed ingest (L29).

`logparse` is carried forward from L22, trimmed to its optimized core: the hand-written parser is now simply *the* parser (the L22 regex oracle + fuzz/bench scaffolding are gone), and `ParseLine` is exported so the ingest handler can count lines individually.

## The package layout

```
lessons/23-http-server/{exercises,solutions}/
├── warmup/middleware/      ← you implement: WithRequestLog + statusRecorder
├── internal/
│   ├── logparse/           carried from L22 (Parse, ParseLine, CountByLevel)
│   └── logstats/           ← you implement: the mutex-guarded Store accumulator
└── cmd/logstats-server/    ← you implement: newRouter + 3 handlers (lifecycle provided)
```

## API reference

```
POST /ingest    {"lines":["2026-01-02T15:04:05 INFO ok", ...]}
                → 200 {"accepted":N,"parsed":M,"failed":K}   (lenient: bad lines counted)
                → 400 on invalid JSON / oversized body
GET  /stats     → 200 {"counts":{"INFO":5,"WARN":2,"ERROR":1},"total":8}
GET  /healthz   → 200 {"status":"ok"}
```

Stats are cumulative since process start. Ingest is lenient — a few malformed log lines are normal, not a client error; only a broken JSON envelope is a 400.

---

## Concept 1 — `net/http` server + ServeMux 1.22 routing

A Go HTTP server is stdlib-only. Since Go 1.22, `ServeMux` understands methods and path patterns:

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /ingest", ingestHandler(store))
mux.HandleFunc("GET /stats", statsHandler(store))
mux.HandleFunc("GET /healthz", healthHandler)
```

- `"POST /ingest"` matches only POST; a GET to `/ingest` gets an **automatic 405** — you don't write that check.
- `"GET /stats/{level}"` captures a segment, read via `r.PathValue("level")`.
- Most-specific pattern wins.

Return `http.Handler` from a builder (`newRouter`) so the real server and the tests share the exact same handler. Use `srv.Serve(listener)` (not `ListenAndServe`) when you want to control the listener — e.g. bind `:0` in a test and read the chosen port.

### Common mistake

The pre-1.22 idiom — register `"/ingest"` then `if r.Method != http.MethodPost { ... 405 ... }` by hand. On Go 1.22+ it's unnecessary and easy to forget on one route, silently accepting the wrong method. Register `"POST /ingest"` and let the mux do it.

---

## Concept 2 — Handlers + request/response

A handler is `func(http.ResponseWriter, *http.Request)`. Everything is explicit:

```go
func ingestHandler(store *logstats.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req ingestRequest
        dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) // 1 MiB cap
        if err := dec.Decode(&req); err != nil {
            http.Error(w, "invalid JSON body", http.StatusBadRequest)
            return
        }
        // parse each line, merge, respond
    }
}
```

- `json.Decoder`/`Encoder` for body in/out. Set `Content-Type` and call `WriteHeader(status)` **before** writing the body.
- `http.MaxBytesReader` caps the body — without it a client can stream gigabytes and OOM you.
- Closures inject dependencies: `ingestHandler(store)` closes over the accumulator — no globals.

Lenient ingest counts malformed lines instead of rejecting them:

```go
delta := map[string]int{}
parsed, failed := 0, 0
for _, line := range req.Lines {
    e, err := logparse.ParseLine(line)
    if err != nil { failed++; continue }
    delta[e.Level]++; parsed++
}
store.Merge(delta)
```

### Common mistake

`json.NewDecoder(r.Body)` with no size limit — a memory-exhaustion DoS. Always wrap with `http.MaxBytesReader`. (And never write the body before `WriteHeader`, or your intended status silently becomes 200.)

---

## Concept 3 — Middleware

Middleware wraps one `http.Handler` and returns another, so cross-cutting concerns compose in layers:

```go
func withRecovery(next http.Handler, logger *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if v := recover(); v != nil {
                logger.Error("panic recovered", "value", v, "path", r.URL.Path)
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

handler := withRequestLog(withRecovery(mux, logger), logger) // chain by nesting
```

To log the status code you must wrap `http.ResponseWriter` — it has no `Status()` method:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}
func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }
```

Default the recorder's status to `http.StatusOK`, since `net/http` sends 200 when a handler writes a body without calling `WriteHeader`.

### Common mistake

Forgetting to call `next.ServeHTTP` (the request does nothing), or trying to read the status without the `statusRecorder` wrapper (there's no way to read it back off `http.ResponseWriter`).

---

## Concept 4 — Structured logging with `slog`

Production logs are *data*. `log/slog` (Go 1.21+) makes structured, leveled, machine-parseable logging a stdlib feature:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
logger.Info("request", "method", "POST", "path", "/ingest", "status", 200)
// {"time":"...","level":"INFO","msg":"request","method":"POST","path":"/ingest","status":200}
```

- `NewJSONHandler` for prod, `NewTextHandler` for dev; filter by `Level`.
- Key/value attributes become structured fields; `logger.With("service","logstats")` stamps attributes on every line (great for request IDs; trace IDs arrive in L28).

### Common mistake

Unstructured interpolation — `logger.Info(fmt.Sprintf("user %s did %s", u, a))` — throws away the structure slog exists for; now you can't filter by user without a regex. And **never log secrets or PII** (tokens, passwords, full bodies): log identifiers, not credentials.

---

## Concept 5 — Designing the service

`net/http` serves **each request in its own goroutine**, so the accumulator is genuine shared concurrent state — the L18 mutex lesson, now load-bearing:

```go
type Store struct {
    mu     sync.Mutex
    counts map[string]int
}
func (s *Store) Merge(delta map[string]int) { s.mu.Lock(); defer s.mu.Unlock(); /* ... */ }
func (s *Store) Snapshot() (map[string]int, int) { /* lock; return a COPY + total */ }
```

`Snapshot` returns a *copy* so `/stats` readers never race the internal map.

Split the lifecycle for testability:

```go
func run(ctx context.Context, addr string, stdout io.Writer) error {
    ln, err := net.Listen("tcp", addr)
    if err != nil { return err }
    return serve(ctx, ln, stdout)
}
```

- Handler tests hand `newRouter(...)` to `httptest.NewServer` — a real server on a random port.
- Lifecycle tests pass a `127.0.0.1:0` listener to `serve` and dial `ln.Addr()` (the L21 pattern).
- `serve` shuts down on `ctx.Done()` via `srv.Shutdown` — a light preview of L26.

### Common mistake

An unguarded shared map. Under concurrent `/ingest` it's a data race — `go test -race` flags it, and production panics with "concurrent map writes" or silently corrupts counts. Every field touched by more than one request goroutine needs a concurrency strategy. **Run `make test-race` on any handler with state.**

---

## Exercise: warm-up — `middleware`

Implement `WithRequestLog(next http.Handler, log *slog.Logger) http.Handler` in `exercises/warmup/middleware/middleware.go` — wrap the handler, capture the status via a `statusRecorder`, log method/path/status/duration.

**Time:** 10-15 minutes.

## Exercise: main — `logstats` accumulator + server

1. **`internal/logstats`** — the mutex-guarded `Store` (`Merge`, `Snapshot` returning a copy).
2. **`cmd/logstats-server`** — `newRouter` + the three handlers (the `run`/`serve` lifecycle, middleware, and `writeJSON` are provided; you fill in routing + handlers). Ingest is lenient; bad JSON is a 400.

**Time:** 45-60 minutes.

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18 — the accumulator is shared state
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/23-http-server/exercises
go test ./...

# Reference solution
go test -v ./lessons/23-http-server/solutions/...
go test -race ./lessons/23-http-server/solutions/...
make test-race

# Run the service
go run ./lessons/23-http-server/solutions/cmd/logstats-server :8080
# (in another terminal)
curl -s -XPOST localhost:8080/ingest \
  -d '{"lines":["2026-01-02T15:04:05 INFO ok","2026-01-02T15:04:06 ERROR boom","bad line"]}'; echo
# {"accepted":3,"parsed":2,"failed":1}
curl -s localhost:8080/stats; echo
# {"counts":{"ERROR":1,"INFO":1},"total":2}
curl -s localhost:8080/healthz; echo
# {"status":"ok"}
```

## Going further

### Read

- **Go blog — "Routing Enhancements for Go 1.22"**: <https://go.dev/blog/routing-enhancements> — the method+pattern `ServeMux`.
- **`log/slog` docs**: <https://pkg.go.dev/log/slog> — handlers, attributes, groups, `With`.
- **Go docs — `net/http`**: <https://pkg.go.dev/net/http> — `Server`, `Handler`, `ServeMux`, `MaxBytesReader`.
- **"How I write HTTP services in Go" (Mat Ryer)** — handler-as-closure + dependency injection, the pattern used here.

### Try

- **Add `GET /stats/{level}`** — return the count for one level using `r.PathValue("level")`; 404 for an unknown level.
- **Add a `-addr` flag** to the server (preview of L26's config lesson) instead of the positional arg.
- **Rate-limit middleware** — wrap the mux in a `withRateLimit`. Staying stdlib-only, build a token bucket from a `time.Ticker` + a buffered channel.
- **Readiness vs liveness** — add `GET /readyz` that returns 503 until startup completes, distinct from `/healthz` liveness.
- **Request IDs** — middleware that generates a request ID, stamps it on the request-scoped `slog` logger via `With`, and echoes it in an `X-Request-ID` response header.

---

> Phase 4 is underway. Next: Lesson 24 — **HTTP clients & resilience**. We build a `logstats-client` that ships log lines to `/ingest` and learn to survive servers that fail, hang, and rate-limit — timeouts, retries with backoff, context propagation, and the circuit-breaker concept.
