<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">23</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>HTTP servers</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Build a real HTTP service with the standard library. Learn <code>net/http</code> + Go 1.22 <code>ServeMux</code> routing, handlers, the middleware pattern, and structured logging with <code>slog</code>. Phase 4 opens: the log aggregator becomes the <strong>logstats service</strong> — clients POST log lines, the server parses and accumulates them, and serves the running counts.</p>
</div>
</div>
</div>

---

## What we'll cover

- **`net/http` server + ServeMux 1.22 routing** — `http.Server`, method+pattern routes.
- **Handlers + request/response** — `http.Handler`, JSON in/out, status codes, body limits.
- **Middleware** — the `func(http.Handler) http.Handler` wrapper; logging + panic recovery.
- **Structured logging with `slog`** — JSON logs, levels, attributes.
- **Designing the service** — the ingest/stats/healthz surface and its shared concurrent state.

---

## The story so far — and Phase 4 begins

Phases 1-3 built the language: types, interfaces, errors, generics, concurrency, profiling. The log aggregator grew from a directory walker (L16) into a measured, hardened batch tool (L22).

**Phase 4 is about production.** We turn that aggregator into a long-lived **service** — the `logstats` service. Instead of walking a directory once, it stays up, accepts log lines over HTTP (`POST /ingest`), parses them with the carried-forward `logparse`, and accumulates per-level counts in shared memory that any client can query (`GET /stats`). Over the next seven lessons this one service gains a resilient client, a gRPC API, configuration, a container, observability, and distributed idempotent ingest.

It all starts with `net/http` — and, refreshingly, **no framework**. The standard library is the framework.

---

## Concept 1: `net/http` server + ServeMux 1.22 routing

### Motivation

A Go HTTP server is a handful of stdlib calls — no Gin, no Echo, no Fiber. `net/http` gives you a production-grade server (HTTP/1.1 + HTTP/2, TLS, timeouts) out of the box. Since Go 1.22, its router (`ServeMux`) understands HTTP methods and path patterns, so you rarely need a third-party router either.

---

### The basics

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /ingest", ingestHandler)
mux.HandleFunc("GET /stats", statsHandler)
mux.HandleFunc("GET /healthz", healthHandler)

srv := &http.Server{Addr: ":8080", Handler: mux}
log.Fatal(srv.ListenAndServe())
```

Go 1.22+ `ServeMux` patterns are `[METHOD ]/path`:

- **Method matching** — `"POST /ingest"` only matches POST. A GET to `/ingest` automatically gets **405 Method Not Allowed** — you don't write that check.
- **Wildcards** — `"GET /stats/{level}"` captures a segment; read it with `r.PathValue("level")`.
- **Precedence** — the most specific pattern wins; no first-match-wins surprises.

Prefer `srv.Serve(listener)` over `ListenAndServe()` when you want to control the listener (e.g. bind `:0` for tests and read the chosen port).

---

### A worked example

The `logstats` router wires three routes and wraps them in middleware:

```go
func newRouter(store *logstats.Store, logger *slog.Logger) http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /ingest", ingestHandler(store))
    mux.HandleFunc("GET /stats", statsHandler(store))
    mux.HandleFunc("GET /healthz", healthHandler)
    return withRequestLog(withRecovery(mux, logger), logger)
}
```

`newRouter` returns an `http.Handler`, so both the real server **and the tests** build the exact same handler — tests just hand it to `httptest.NewServer`.

---

### Common mistake

**The pre-1.22 idiom, still copy-pasted everywhere:**

```go
// the old way — verbose and error-prone
mux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {            // ❌ manual method check
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // ...
})
```

On Go 1.22+ this is unnecessary. Register `"POST /ingest"` and the mux returns 405 for you. Writing the manual check is a sign of pre-1.22 muscle memory — and it's easy to forget on one route and silently accept the wrong method.

---

### Recap

- `net/http` is a production server; no framework needed.
- Go 1.22 `ServeMux` matches method + pattern (`"POST /ingest"`) and gives automatic 405.
- `{wildcard}` segments via `r.PathValue`.
- Return `http.Handler` from a builder so server and tests share it; use `srv.Serve(ln)` for test control.

---

## Concept 2: Handlers + request/response

### Motivation

A handler is where a request becomes a response. In Go it's tiny: a function with the signature `func(http.ResponseWriter, *http.Request)`. Everything — JSON decoding, status codes, headers — is explicit. That explicitness is the lesson: you see exactly what crosses the wire.

---

### The basics

```go
type ingestRequest struct {
    Lines []string `json:"lines"`
}

func ingestHandler(store *logstats.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req ingestRequest
        dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)) // 1 MiB cap
        if err := dec.Decode(&req); err != nil {
            http.Error(w, "invalid JSON body", http.StatusBadRequest)
            return
        }
        // ... process req.Lines ...
        writeJSON(w, http.StatusOK, ingestResponse{ /* ... */ })
    }
}
```

Three things:

- **`json.Decoder`** reads the request body; **`json.Encoder`** writes the response. Set `Content-Type: application/json` and call `WriteHeader(status)` **before** writing the body.
- **`http.MaxBytesReader`** caps the body — without it, a client can stream gigabytes and exhaust memory.
- **Closures carry dependencies.** `ingestHandler(store)` returns a handler closed over `store` — the Go way to "inject" the accumulator without globals.

The `writeJSON` helper centralizes the header + status + encode:

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}
```

---

### A worked example

`logstats` ingest is **lenient** — log streams are messy, so malformed lines are counted, not rejected:

```go
delta := map[string]int{}
parsed, failed := 0, 0
for _, line := range req.Lines {
    e, err := logparse.ParseLine(line)
    if err != nil {
        failed++
        continue
    }
    delta[e.Level]++
    parsed++
}
store.Merge(delta)
writeJSON(w, http.StatusOK, ingestResponse{Accepted: len(req.Lines), Parsed: parsed, Failed: failed})
```

A request of 4 lines (3 valid, 1 garbage) returns `{"accepted":4,"parsed":3,"failed":1}` with status 200. Only a broken JSON *envelope* is a 400 — a few bad log lines are normal, not a client error.

---

### Common mistake

**Decoding an unbounded body.** `json.NewDecoder(r.Body)` with no limit lets a malicious or buggy client send an arbitrarily large body and OOM your server:

```go
json.NewDecoder(r.Body).Decode(&req)   // ❌ no size limit → memory-exhaustion DoS
```

Always wrap with `http.MaxBytesReader(w, r.Body, max)`. (Secondary mistake: writing the body before `WriteHeader`, which locks in a 200 and makes your intended status a silent no-op.)

---

### Recap

- A handler is `func(http.ResponseWriter, *http.Request)`; `http.HandlerFunc` adapts it.
- `json.Decoder`/`Encoder` for request/response; set Content-Type + status before the body.
- **Cap request bodies** with `http.MaxBytesReader`.
- Closures inject dependencies (the `store`) — no globals.

---

## Concept 3: Middleware

### Motivation

Cross-cutting concerns — logging, auth, recovery, rate limiting — shouldn't be copy-pasted into every handler. Middleware factors them out: a function that wraps one `http.Handler` and returns another, so you compose behavior in layers around your routes.

---

### The basics

The shape is always the same:

```go
func withRecovery(next http.Handler, logger *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if v := recover(); v != nil {
                logger.Error("panic recovered", "value", v, "path", r.URL.Path)
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)   // call the wrapped handler
    })
}
```

You **chain** them by nesting — innermost runs last:

```go
handler := withRequestLog(withRecovery(mux, logger), logger)
// request flow: withRequestLog → withRecovery → mux → handler
```

To log the **status code**, you must wrap `http.ResponseWriter`, because it doesn't expose the code after the fact:

```go
type statusRecorder struct {
    http.ResponseWriter
    status int
}
func (s *statusRecorder) WriteHeader(code int) {
    s.status = code
    s.ResponseWriter.WriteHeader(code)
}
```

---

### A worked example

The logging middleware times the request and records the status via the wrapper:

```go
func withRequestLog(next http.Handler, logger *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(rec, r)
        logger.Info("request",
            "method", r.Method, "path", r.URL.Path,
            "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
    })
}
```

`status: http.StatusOK` is the default because `net/http` sends 200 when a handler writes a body without calling `WriteHeader` — so a handler that just `w.Write(...)`s is correctly logged as 200.

---

### Common mistake

**Forgetting to call `next.ServeHTTP`** (the request silently does nothing), or **recovering from a panic but still letting a 200 through** (the wrapped handler may have written partial output before panicking). A subtler one: trying to read `w`'s status without the `statusRecorder` wrapper — there's no `w.Status()` method; you *must* intercept `WriteHeader`.

---

### Recap

- Middleware is `func(http.Handler) http.Handler` — wrap and return.
- Chain by nesting; innermost handler runs last.
- Capture the status code by wrapping `http.ResponseWriter` and overriding `WriteHeader`.
- Recovery middleware turns a handler panic into a 500 instead of crashing the server.

---

## Concept 4: Structured logging with `slog`

### Motivation

`log.Printf("request %s %s %d", ...)` produces strings humans skim and machines can't query. Production logs are *data* — you grep them in Loki, filter them in CloudWatch, alert on them in Datadog. Go 1.21's `log/slog` makes structured, leveled, machine-parseable logging a stdlib feature.

---

### The basics

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

logger.Info("request", "method", "POST", "path", "/ingest", "status", 200)
// {"time":"...","level":"INFO","msg":"request","method":"POST","path":"/ingest","status":200}
```

- **Handlers** decide the format: `NewJSONHandler` (production) or `NewTextHandler` (dev).
- **Key/value attributes** after the message become structured fields. Prefer typed helpers (`slog.Int`, `slog.String`) in hot paths to avoid the `any` boxing.
- **`logger.With("service", "logstats")`** returns a child logger that stamps those attributes on every line — perfect for request-scoped context (request ID, trace ID — the latter arrives in L28).
- **Levels:** `Debug`/`Info`/`Warn`/`Error`, filtered by the handler's `Level`.

---

### A worked example

The service builds one JSON logger and threads it through the middleware:

```go
func serve(ctx context.Context, ln net.Listener, stdout io.Writer) error {
    logger := slog.New(slog.NewJSONHandler(stdout, nil))
    store := logstats.NewStore()
    srv := &http.Server{Handler: newRouter(store, logger)}
    // ...
    logger.Info("listening", "addr", ln.Addr().String())
    // ...
}
```

Every request line, every recovered panic, and the startup line are all JSON on the same stream — uniform, queryable, ready to ship to a log aggregator (perhaps even one built like ours).

---

### Common mistake

**Unstructured interpolation** (`logger.Info(fmt.Sprintf("user %s did %s", u, a))`) throws away the structure slog exists to provide — now you can't filter by user without a regex. And **logging secrets/PII** (tokens, passwords, full request bodies) is a compliance incident waiting to happen: log identifiers, not credentials.

---

### Recap

- `slog` = structured, leveled, machine-parseable logging in the stdlib.
- `NewJSONHandler` for prod, `NewTextHandler` for dev; filter by `Level`.
- Key/value attributes, not interpolated strings; `With` for request-scoped context.
- Never log secrets or PII.

---

## Concept 5: Designing the service

### Motivation

The endpoints are easy; the design decisions are what make it a *service*: what state does it hold, how is that state shared safely across concurrent requests, and how do you test the whole thing without binding a real port?

---

### The basics — the surface

| Endpoint | Purpose | Response |
|---|---|---|
| `POST /ingest` | submit log lines | `{"accepted":N,"parsed":M,"failed":K}` |
| `GET /stats` | read level counts | `{"counts":{...},"total":N}` |
| `GET /healthz` | liveness probe | `{"status":"ok"}` |

The **shared state** is the accumulator — and it's touched by every concurrent `/ingest`:

```go
type Store struct {
    mu     sync.Mutex
    counts map[string]int
}
func (s *Store) Merge(delta map[string]int) { s.mu.Lock(); defer s.mu.Unlock(); /* ... */ }
func (s *Store) Snapshot() (map[string]int, int) { /* lock; return a COPY */ }
```

`net/http` serves **each request in its own goroutine**. So the accumulator is genuine shared concurrent state — exactly the L18 lesson, now load-bearing. `Snapshot` returns a *copy* so `/stats` readers never race the internal map.

---

### A worked example — testability

Split the lifecycle so tests don't need a fixed port:

```go
func run(ctx context.Context, addr string, stdout io.Writer) error {
    ln, err := net.Listen("tcp", addr)
    if err != nil { return err }
    return serve(ctx, ln, stdout)
}
```

- **Handler tests** hand `newRouter(...)` to `httptest.NewServer` — a real server on a random port, no lifecycle.
- **Lifecycle tests** pass a `127.0.0.1:0` listener to `serve` and dial `ln.Addr()` (the L21 pattern).
- `serve` shuts down on `ctx.Done()` via `srv.Shutdown` — a light preview of L26's graceful shutdown.

```go
func TestConcurrentIngestRace(t *testing.T) {
    srv := httptest.NewServer(testRouter())
    defer srv.Close()
    // 50 concurrent POSTs → run under `go test -race`
}
```

---

### Common mistake

**An unguarded shared map.** The tempting first cut:

```go
type Store struct { counts map[string]int }   // ❌ no mutex
func (s *Store) Merge(d map[string]int) { for k, n := range d { s.counts[k] += n } }
```

Under concurrent `/ingest` this is a data race — `go test -race` flags it, and in production it corrupts counts or panics with "concurrent map writes." The fix is the mutex (or `sync/atomic` for single counters). **Every field touched by more than one request goroutine needs a concurrency strategy.** Make `make test-race` a reflex for any handler with state.

---

### Recap

- The service surface: `POST /ingest`, `GET /stats`, `GET /healthz`.
- `net/http` runs each request in its own goroutine → shared state needs a mutex (L18 made real).
- `Snapshot` returns a copy so readers don't race writers.
- Split `run`/`serve` + use `httptest` so the whole service is testable without fixed ports.

---

## Practice

### Warm-up

In `exercises/warmup/middleware/`, implement `WithRequestLog(next http.Handler, log *slog.Logger) http.Handler` — wrap the handler, capture the status via a `statusRecorder`, and log method/path/status/duration.

```bash
cd lessons/23-http-server/exercises
go test ./warmup/middleware/...
```

---

### Main

1. **`internal/logstats`** — implement the mutex-guarded `Store` (`Merge`, `Snapshot` returning a copy).
2. **`cmd/logstats-server`** — implement `newRouter` + the three handlers (the `run`/`serve` lifecycle and middleware are provided). Ingest is lenient; bad JSON is a 400.

```bash
cd lessons/23-http-server/exercises
go test ./...
go test -race ./...
make test-race      # daily habit (since L18)

# run it
go run ./cmd/logstats-server :8080
curl -s -XPOST localhost:8080/ingest -d '{"lines":["2026-01-02T15:04:05 INFO ok","bad"]}'; echo
curl -s localhost:8080/stats; echo
curl -s localhost:8080/healthz; echo
```

---

## Closing thought

You just built a production-shaped HTTP service with nothing but the standard library: a router that understands methods, JSON handlers with body limits, composable middleware, structured logs, and concurrency-safe shared state — all testable without binding a real port.

This is the Go philosophy in miniature. The stdlib is not a starter kit you outgrow; it's the foundation real services ship on. Frameworks add conventions on top, but everything underneath is what you wrote today. When you read Gin or Chi source, you'll recognize all of it — they're `http.Handler` wrappers too.

---

## What we learned

- **`net/http` + ServeMux 1.22:** method+pattern routing, automatic 405, `{wildcard}` paths — no framework.
- **Handlers:** `json.Decoder`/`Encoder`, status-before-body, `http.MaxBytesReader`, dependency injection via closures.
- **Middleware:** `func(http.Handler) http.Handler`, chaining, the `statusRecorder` trick, panic recovery.
- **`slog`:** structured JSON logs, levels, attributes, `With`; never log secrets.
- **Service design:** each request is a goroutine → mutex-guarded shared state (L18 made real); `run`/`serve` + `httptest` for testability.

---

## Up next

Lesson 24 — **HTTP clients & resilience**. The other side of the wire: a `logstats-client` that ships log lines to `/ingest` with timeouts, retries + exponential backoff, context propagation, and the circuit-breaker concept. We learn to call services that fail, hang, and rate-limit — because in production, they will.
