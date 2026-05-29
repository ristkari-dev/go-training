<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">25</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 4 — Production &amp; Distributed</div>
<h1>gRPC</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Give the <code>logstats</code> service a typed, streaming RPC face alongside its HTTP face. Learn protobuf as an IDL, the codegen workflow, unary and client-streaming RPCs, and interceptors (gRPC's middleware). This is the course's first third-party dependency — <code>google.golang.org/grpc</code> — introduced deliberately, because you can't honestly teach gRPC without it.</p>
</div>
</div>
</div>

---

## What we'll cover

- **Protobuf IDL + type system** — messages, services, field numbers, wire compatibility.
- **Codegen workflow** — `protoc` + plugins, committed `*.pb.go`, `make proto`.
- **Unary RPC** — request/response, like a typed function call across the network.
- **Client-streaming RPC** — stream many messages, get one reply.
- **Interceptors + gRPC vs REST** — middleware for RPCs, and when to choose which.

---

## The story so far — a second face on the same service

L23/L24 gave `logstats` an HTTP face (`POST /ingest`, `GET /stats`) and a resilient client. **L25 adds a gRPC face** — `Ingest` (client-streaming) and `GetStats` (unary) — over the **same `logstats.Store`**. The HTTP handlers and the gRPC service call the identical domain core (`logparse.ParseLine` + the mutex-guarded accumulator). The transport is just a face; the logic underneath is shared.

gRPC trades REST's human-readable, ubiquitous, cacheable HTTP/JSON for a **typed contract** (the `.proto`), **code generation** (no hand-written client/server boilerplate, no field-name typos), **streaming** (client/server/bidirectional), and **HTTP/2** performance. It's the default for internal service-to-service APIs at scale.

And it's a milestone: after 24 lessons of pure standard library, we add our first real dependency. We do it the right way — pinned, committed, minimal.

---

## Concept 1: Protobuf IDL + type system

### Motivation

A REST+JSON API's "contract" lives in documentation and hope: the server and client independently agree on field names and types, and drift silently. gRPC's contract is a **file** — the `.proto` — that both sides generate code from. It's the single source of truth, machine-checked.

---

### The basics

```proto
syntax = "proto3";
package logstats.v1;
option go_package = ".../lessons/25-grpc/<tree>/proto/logstatspb"; // <tree> = exercises | solutions

service LogStats {
  rpc Ingest(stream LogLine) returns (IngestSummary);   // client-streaming
  rpc GetStats(StatsRequest) returns (StatsReply);       // unary
}

message LogLine     { string line = 1; }
message IngestSummary { int32 accepted = 1; int32 parsed = 2; int32 failed = 3; }
message StatsRequest  {}
message StatsReply    { map<string, int32> counts = 1; int32 total = 2; }
```

- **`message`** = a struct; **`service`** = a set of `rpc`s.
- **Field numbers** (`= 1`, `= 2`) are the wire identity — they, not the names, are encoded on the wire.
- Scalars (`string`, `int32`, `bool`…), `map<K,V>`, `repeated T` (slices). proto3 fields are non-optional with zero-value defaults.
- `go_package` tells the Go generator where the generated package lives.

---

### A worked example

`StatsReply` carries `map<string, int32> counts` — the same shape the HTTP `/stats` returns as JSON. The proto compiles to a Go struct with a `GetCounts() map[string]int32` accessor; the server fills it from `logstats.Store.Snapshot()`. One contract, two faces.

---

### Common mistake

**Changing or reusing a field number.** The wire format keys on numbers, not names. Renaming `line` to `text` (keeping `= 1`) is wire-compatible — old and new peers still interop. But changing `line = 1` to `line = 2`, or reusing a retired number for a new field, **silently corrupts** data exchanged with peers compiled against the old number. Rule: field numbers are forever; add new fields with new numbers, never recycle.

---

### Recap

- The `.proto` is the machine-checked contract; both sides generate from it.
- `message` → struct, `service` → RPCs; field **numbers** are the wire identity.
- proto3 has zero-value defaults; `map`/`repeated` cover collections.
- Never change or reuse a field number.

---

## Concept 2: The codegen workflow

### Motivation

You don't write gRPC client/server code by hand — you generate it from the `.proto`. `protoc` (the protobuf compiler) plus two Go plugins emit the message structs, the client stub, and the server interface. Your job is to implement the server interface and call the client stub.

---

### The basics

```bash
protoc --go_out=. --go_opt=module=<modpath> \
       --go-grpc_out=. --go-grpc_opt=module=<modpath> \
       proto/logstats.proto
```

- **`protoc-gen-go`** generates `logstats.pb.go` — the message types + accessors.
- **`protoc-gen-go-grpc`** generates `logstats_grpc.pb.go` — `LogStatsClient`, `LogStatsServer`, `RegisterLogStatsServer`, the stream types.
- **`module=<modpath>`** routes output to the dir implied by `go_package` minus the module prefix.

We wrap this in a `make proto` target and **commit the generated `*.pb.go`**, so building the project (and CI, and students) needs no `protoc` — only `go build`. Regenerate only when the `.proto` changes.

---

### A worked example

The lesson's `make proto` compiles each `.proto` in its own invocation (the exercises and solutions trees reuse the same proto package names, which would collide if compiled as one unit). The committed output: `proto/logstatspb/{logstats.pb.go, logstats_grpc.pb.go}` — ~450 lines you never wrote or maintain by hand.

Tool versions are pinned (in this lesson: `protoc-gen-go` v1.36.6, `protoc-gen-go-grpc` v1.5.1) so regeneration is reproducible.

---

### Common mistake

**Hand-editing generated files.** They carry `// Code generated ... DO NOT EDIT.` for a reason — your edits vanish on the next `make proto`. Need different behavior? Change the `.proto` and regenerate, or wrap the generated type in your own. (Second mistake: *not* committing the generated code, so a teammate's `go build` fails because they don't have `protoc` or the right plugins.)

---

### Recap

- `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc` generate messages + client + server interface.
- Commit the `*.pb.go`; build needs no `protoc`. Pin plugin versions for reproducibility.
- Never hand-edit generated files — change the `.proto`.

---

## Concept 3: Unary RPC

### Motivation

The simplest RPC: one request, one response — a typed function call that happens to cross the network. `GetStats` is ours. The generated code makes it feel local: the client calls a method, the server implements one.

---

### The basics

The generator emits a server interface; you implement it:

```go
type Server struct {
    pb.UnimplementedLogStatsServer // forward-compat: embed this
    store *logstats.Store
}

func (s *Server) GetStats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsReply, error) {
    counts, total := s.store.Snapshot()
    out := make(map[string]int32, len(counts))
    for k, v := range counts {
        out[k] = int32(v)
    }
    return &pb.StatsReply{Counts: out, Total: int32(total)}, nil
}
```

The client side is a generated stub:

```go
client := pb.NewLogStatsClient(conn)
reply, err := client.GetStats(ctx, &pb.StatsRequest{})
// reply.GetTotal(), reply.GetCounts()
```

Errors cross the wire as gRPC `status` codes (`codes.NotFound`, `codes.InvalidArgument`, …) — richer than HTTP status alone.

---

### A worked example

`GetStats` reads the shared `logstats.Store` — the very accumulator the HTTP `/stats` handler reads. Two transports, one source of truth. The `ctx` carries deadlines/cancellation across the wire, exactly like the HTTP client context from L24.

---

### Common mistake

**Not embedding `pb.UnimplementedLogStatsServer`.** It provides default "unimplemented" methods so your `Server` satisfies the interface even as new RPCs are added to the `.proto`. Omit it, and the day someone adds a third RPC, every server in the codebase fails to compile until each is updated. Embedding it makes new RPCs default to a clean `Unimplemented` error instead — forward compatibility for free.

---

### Recap

- Unary = one request → one response; the generated stub makes it feel like a local call.
- Implement the server interface; embed `Unimplemented…Server` for forward compatibility.
- `ctx` propagates deadlines/cancellation; errors are typed `status` codes.

---

## Concept 4: Client-streaming RPC

### Motivation

`GetStats` is one-shot. But ingestion is a *stream* — a client tails a log and sends lines continuously. gRPC models this natively with streaming RPCs. `Ingest` is **client-streaming**: the client sends many `LogLine`s, the server replies once with an `IngestSummary`.

---

### The basics

gRPC has four shapes: **unary**, **server-streaming** (one request → many responses), **client-streaming** (many → one), and **bidirectional**. `stream` on the request type marks client-streaming:

```proto
rpc Ingest(stream LogLine) returns (IngestSummary);
```

Server: loop `Recv()` until `io.EOF`, then `SendAndClose` once:

```go
func (s *Server) Ingest(stream pb.LogStats_IngestServer) error {
    delta := map[string]int{}
    var accepted, parsed, failed int32
    for {
        line, err := stream.Recv()
        if err == io.EOF {
            s.store.Merge(delta)
            return stream.SendAndClose(&pb.IngestSummary{Accepted: accepted, Parsed: parsed, Failed: failed})
        }
        if err != nil {
            return err
        }
        accepted++
        if e, perr := logparse.ParseLine(line.GetLine()); perr == nil {
            delta[e.Level]++; parsed++
        } else {
            failed++
        }
    }
}
```

Client: `Send()` in a loop, then `CloseAndRecv()` for the summary:

```go
stream, _ := client.Ingest(ctx)
for sc.Scan() { stream.Send(&pb.LogLine{Line: sc.Text()}) }
summary, _ := stream.CloseAndRecv()
```

---

### A worked example

`Ingest` mirrors HTTP `/ingest` exactly — lenient counting (malformed lines increment `failed`, not rejected) over the same `logparse.ParseLine` + `Store.Merge`. The difference is the *shape*: instead of batching lines into one JSON body, the client streams them one message at a time and the server folds them as they arrive.

---

### Common mistake

**Not terminating the stream.** The server's `Recv()` loop ends only on `io.EOF` (sent when the client calls `CloseSend`/`CloseAndRecv`). Forget `CloseAndRecv` on the client and the server blocks in `Recv()` forever; forget the `io.EOF` check on the server and you never send the summary. The terminal handshake — client closes, server `SendAndClose` — is mandatory.

---

### Recap

- Four RPC shapes; `stream` on a type marks streaming. `Ingest` is client-streaming (many → one).
- Server: `Recv()` to `io.EOF`, then `SendAndClose`. Client: `Send()` loop, then `CloseAndRecv`.
- Same domain logic as HTTP `/ingest`; only the transport shape differs.

---

## Concept 5: Interceptors + when gRPC beats REST

### Motivation

Cross-cutting concerns — logging, auth, metrics, tracing — shouldn't live in every RPC method, exactly as they shouldn't live in every HTTP handler (L23's middleware). gRPC's equivalent is the **interceptor**.

---

### The basics

A unary interceptor wraps every unary call; a stream interceptor wraps every streaming call:

```go
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        resp, err := handler(ctx, req)
        logger.Info("unary", "method", info.FullMethod, "err", err)
        return resp, err
    }
}

srv := grpc.NewServer(
    grpc.UnaryInterceptor(LoggingUnaryInterceptor(logger)),
    grpc.StreamInterceptor(LoggingStreamInterceptor(logger)),
)
```

`info.FullMethod` is `/logstats.v1.LogStats/GetStats` — the routed RPC. Interceptors chain, just like HTTP middleware.

---

### When gRPC beats REST

| gRPC | REST + JSON |
|---|---|
| Typed contract + codegen (no drift) | Hand-agreed schema; drifts |
| Streaming (client/server/bidi) | Request/response (SSE/WebSocket bolt-ons) |
| HTTP/2 multiplexing, binary, compact | HTTP/1.1-friendly, text, verbose |
| Hard to call from a browser/curl | Trivial from any client; cacheable |

**gRPC** shines for internal service-to-service APIs, streaming, and polyglot fleets. **REST** shines for public/browser-facing APIs, debuggability, and ecosystem ubiquity. Our service offers both faces — a real pattern (gRPC internally, a REST gateway at the edge).

---

### Common mistake

**Cross-cutting logic copy-pasted into every method** instead of an interceptor — the same anti-pattern L23 flagged for HTTP handlers. The other one: **reaching for gRPC for a browser-facing public API**, where you'll fight CORS, lack of native browser support (needs grpc-web + a proxy), and un-curl-able payloads. Match the transport to the consumer.

---

### Recap

- Interceptors are gRPC middleware — unary + stream — for logging, auth, metrics; they chain.
- gRPC: typed, streaming, fast, internal. REST: debuggable, ubiquitous, browser/public.
- Don't put cross-cutting logic in every method; don't expose gRPC directly to browsers.

---

## Practice

### Warm-up

In `exercises/warmup/greeter/`, implement the unary `Greet` server method (the `.proto` + generated `greeterpb` are provided): return `"Hello, <name>!"`. Test it over `bufconn` (in-process gRPC — no real port).

```bash
cd lessons/25-grpc/exercises
go test ./warmup/greeter/...
```

### Main

1. **`internal/grpcsrv`** — implement `GetStats` (unary), `Ingest` (client-streaming, lenient counting), and the logging unary/stream interceptors, over the shared `logstats.Store`.
2. **`cmd/logstats-grpc-server` + `cmd/logstats-grpc-client`** (provided) — study the wiring: `grpc.NewServer` + interceptors + `RegisterLogStatsServer`; the client streams stdin via `Ingest` then calls `GetStats`.

```bash
cd lessons/25-grpc/exercises
go test ./...
go test -race ./...
make test-race

# regenerate code if you change a .proto
make proto

# end-to-end over a real port
go run ./cmd/logstats-grpc-server 127.0.0.1:9090 &
printf '2026-01-02T15:04:05 INFO ok\nbad\n' | go run ./cmd/logstats-grpc-client -addr=127.0.0.1:9090
```

---

## Closing thought

gRPC looks like a lot — a new IDL, a compiler, plugins, generated code, four streaming shapes. But strip it down and it's the same idea as everything in this phase: a typed contract, generated boilerplate so you write only the logic, and middleware for the cross-cutting parts. The `Ingest` server method is *the same code* as the HTTP `ingestHandler` — `ParseLine`, count, `Merge` — wearing a different collar.

That's the lesson worth keeping: **transport is a face over domain logic.** HTTP and gRPC are two doors into the same room. Design the room — the `logstats.Store`, the parser — well, and adding a door is an afternoon. We just added our first external dependency to do it, pinned and committed and minimal, because some doors are worth not building by hand.

---

## What we learned

- **Protobuf**: the `.proto` is the machine-checked contract; field numbers are the wire identity (never reuse them).
- **Codegen**: `protoc` + plugins generate messages/client/server; commit the `*.pb.go`, pin the plugins, `make proto` to regenerate.
- **Unary RPC**: implement the server interface (embed `Unimplemented…` for forward compat); the client stub feels local; errors are typed `status` codes.
- **Client-streaming**: `Recv()` to `io.EOF` then `SendAndClose`; client `Send()` loop then `CloseAndRecv`; same domain core as HTTP `/ingest`.
- **Interceptors**: gRPC middleware (unary + stream); gRPC for internal/streaming, REST for public/browser.

---

## Up next

Lesson 26 — **Configuration, secrets & graceful shutdown**. The service learns to be configured (flags > env > file precedence), to keep secrets out of logs, and to shut down gracefully — draining in-flight HTTP requests and gRPC calls on SIGTERM with `http.Server.Shutdown` and `grpc.GracefulStop`. The pieces that turn "runs on my machine" into "runs in production."
