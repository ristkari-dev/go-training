# Lesson 25: gRPC

## What you'll learn

By the end of this lesson you can:

- Define a service contract in **protobuf** and understand the wire-compatibility rules.
- Generate Go client/server code with **`protoc`** + plugins, and manage committed generated code via `make proto`.
- Implement and call a **unary RPC** (request → response).
- Implement and call a **client-streaming RPC** (many messages → one reply).
- Add **interceptors** (gRPC middleware) and reason about **gRPC vs REST**.

This is the third Phase 4 lesson and the course's **first third-party dependencies** — `google.golang.org/grpc` + `google.golang.org/protobuf`, introduced deliberately.

## What's different from L24

L23/L24 gave `logstats` an HTTP face and a resilient client. **L25 adds a gRPC face** — `Ingest` (client-streaming) + `GetStats` (unary) — over the **same `logstats.Store`**. The gRPC service calls the identical domain core (`logparse.ParseLine` + the mutex-guarded accumulator) as the HTTP handlers. Transport is a face over domain logic; both doors open into the same room.

After 24 stdlib-only lessons, this is the first time we add external dependencies. We do it the right way: **pinned, committed, minimal** (see "Dependencies" below).

## The package layout

```
lessons/25-grpc/{exercises,solutions}/
├── warmup/greeter/         ← you implement: Greet (unary); proto + greeterpb provided
│   ├── greeter.proto, greeterpb/*.pb.go (generated, committed)
├── proto/
│   ├── logstats.proto      the LogStats contract (provided)
│   └── logstatspb/*.pb.go   generated, committed
├── internal/
│   ├── logparse/           carried from L24 (the parser)
│   ├── logstats/           carried from L24 (the shared accumulator)
│   └── grpcsrv/            ← you implement: service impl + interceptors
└── cmd/
    ├── logstats-server/        carried from L23/L24 (the HTTP face — still here)
    ├── logstats-grpc-server/   provided: registers grpcsrv over TCP
    └── logstats-grpc-client/   provided: streams stdin via Ingest, then GetStats
```

## Dependencies (the boundary shifts here)

```
google.golang.org/grpc      v1.73.0    // newest grpc that keeps the module at go 1.23
google.golang.org/protobuf  v1.36.11
```

- **Why grpc v1.73.0, not latest?** grpc v1.76+ declares `go 1.24` and v1.81 declares `go 1.25`; v1.73.0 is the newest that declares `go 1.23`, so adding it keeps the course module's `go 1.23` directive intact. (A clean `go mod tidy` with v1.73.0 leaves `go 1.23`, four small indirect deps, ~34 go.sum lines.)
- **Generated code is committed**, so `go build`, CI, and students need no `protoc` — only `go build`. Regenerate with `make proto` only when a `.proto` changes.
- **Tool versions** (for reproducible regen): `protoc-gen-go` v1.36.6, `protoc-gen-go-grpc` v1.5.1, `protoc` libprotoc 35.x.
- `.golangci.yml` excludes `*.pb.go` from linting (generated code isn't ours to lint).

## The contract (`proto/logstats.proto`)

```proto
service LogStats {
  rpc Ingest(stream LogLine) returns (IngestSummary);   // client-streaming
  rpc GetStats(StatsRequest) returns (StatsReply);       // unary
}
message LogLine       { string line = 1; }
message IngestSummary { int32 accepted = 1; int32 parsed = 2; int32 failed = 3; }
message StatsRequest  {}
message StatsReply    { map<string, int32> counts = 1; int32 total = 2; }
```

`Ingest` mirrors HTTP `/ingest` (many lines → one summary); `GetStats` mirrors `/stats`.

---

## Concept 1 — Protobuf IDL + type system

The `.proto` is the machine-checked contract both sides generate from — no documentation drift. `message` = struct, `service` = a set of `rpc`s; field **numbers** (`= 1`) are the wire identity (names are not encoded). proto3 fields have zero-value defaults; `map<K,V>` and `repeated T` cover collections.

### Common mistake

Changing or reusing a field number. Renaming a field (keeping its number) is wire-compatible; changing its number, or recycling a retired number, silently corrupts data exchanged with peers compiled against the old number. **Field numbers are forever.**

---

## Concept 2 — The codegen workflow

`protoc` + `protoc-gen-go` (messages) + `protoc-gen-go-grpc` (client/server/stream types) generate the code:

```bash
make proto   # wraps: protoc --go_out=. --go_opt=module=<mod> --go-grpc_out=. --go-grpc_opt=module=<mod> <proto>
```

Each `.proto` is compiled in its own invocation (the exercises/solutions trees reuse the same proto package names, which would collide in one unit). The committed `*.pb.go` (~450 lines you never hand-write) means the build needs no `protoc`.

### Common mistake

Hand-editing generated files (lost on the next `make proto` — they say `DO NOT EDIT`), or *not committing* them so teammates' `go build` fails. Change the `.proto` and regenerate; commit the result.

---

## Concept 3 — Unary RPC

One request → one response — a typed function call across the network. Implement the generated server interface; call the generated client stub:

```go
func (s *Server) GetStats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsReply, error) {
    counts, total := s.store.Snapshot()
    out := make(map[string]int32, len(counts))
    for k, v := range counts { out[k] = int32(v) }
    return &pb.StatsReply{Counts: out, Total: int32(total)}, nil
}

// client:
reply, err := pb.NewLogStatsClient(conn).GetStats(ctx, &pb.StatsRequest{})
```

`ctx` carries deadlines/cancellation across the wire; errors are typed gRPC `status` codes.

### Common mistake

Not embedding `pb.UnimplementedLogStatsServer` in your server struct. It supplies defaults so your server still satisfies the interface when new RPCs are added to the `.proto` — without it, adding an RPC breaks every server at compile time. Embedding it = forward compatibility for free.

---

## Concept 4 — Client-streaming RPC

gRPC has four shapes: unary, server-streaming, client-streaming, bidirectional. `Ingest` is client-streaming (`stream` on the request): the client sends many `LogLine`s, the server replies once.

```go
// server: Recv until io.EOF, then SendAndClose
for {
    line, err := stream.Recv()
    if err == io.EOF {
        s.store.Merge(delta)
        return stream.SendAndClose(&pb.IngestSummary{Accepted: accepted, Parsed: parsed, Failed: failed})
    }
    if err != nil { return err }
    accepted++
    if e, perr := logparse.ParseLine(line.GetLine()); perr == nil { delta[e.Level]++; parsed++ } else { failed++ }
}

// client: Send loop, then CloseAndRecv
stream, _ := client.Ingest(ctx)
for sc.Scan() { stream.Send(&pb.LogLine{Line: sc.Text()}) }
summary, _ := stream.CloseAndRecv()
```

Lenient counting (malformed lines increment `failed`) over the same `logparse.ParseLine` + `Store.Merge` as HTTP `/ingest` — only the *shape* differs (streamed messages vs one JSON batch).

### Common mistake

Not terminating the stream: the server's `Recv()` loop ends only on `io.EOF` (sent by the client's `CloseAndRecv`/`CloseSend`). Forget it and the server blocks forever; forget the `io.EOF` check and the summary never sends.

---

## Concept 5 — Interceptors + when gRPC beats REST

Interceptors are gRPC's middleware — one for unary calls, one for streams — for logging/auth/metrics, registered on the server and chained:

```go
srv := grpc.NewServer(
    grpc.UnaryInterceptor(LoggingUnaryInterceptor(logger)),
    grpc.StreamInterceptor(LoggingStreamInterceptor(logger)),
)
```

**gRPC** wins for internal service-to-service APIs, streaming, and polyglot fleets (typed contract, codegen, HTTP/2). **REST** wins for public/browser-facing APIs, debuggability, and ubiquity (curl, caching, no proxy). Our service offers both — gRPC internally, REST at the edge.

### Common mistake

Cross-cutting logic copy-pasted into every method instead of an interceptor (the L23 HTTP-middleware lesson again); or exposing gRPC directly to browsers (CORS pain, needs grpc-web + a proxy, un-curl-able). Match the transport to the consumer.

---

## Exercise: warm-up — `greeter`

Implement the unary `Greet` method in `exercises/warmup/greeter/greeter.go` (`"Hello, <name>!"`). The `.proto` + generated `greeterpb` are provided. Test it over `bufconn` (in-process gRPC).

**Time:** 10-15 minutes.

## Exercise: main — `grpcsrv`

Implement, in `exercises/internal/grpcsrv/server.go`:
1. **`GetStats`** (unary) — snapshot the store into a `StatsReply`.
2. **`Ingest`** (client-streaming) — `Recv` to `io.EOF`, lenient parse+count, `Merge`, `SendAndClose`.
3. **`LoggingUnaryInterceptor` + `LoggingStreamInterceptor`** — call the handler, log `info.FullMethod` + err.

The cmd binaries (`logstats-grpc-server`, `logstats-grpc-client`) are provided — study how they wire `grpc.NewServer` + interceptors + the client stream.

**Time:** 45-60 minutes.

---

## Daily habits

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18 — the accumulator is shared state
make proto        # only when a .proto changes
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/25-grpc/exercises
go test ./...

# Reference solution
go test -v ./lessons/25-grpc/solutions/...
go test -race ./lessons/25-grpc/solutions/...
make test-race

# End-to-end: gRPC server + client over a real port
go build -o /tmp/grpc-server ./lessons/25-grpc/solutions/cmd/logstats-grpc-server
go build -o /tmp/grpc-client ./lessons/25-grpc/solutions/cmd/logstats-grpc-client
/tmp/grpc-server 127.0.0.1:9090 &
printf '2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n' \
  | /tmp/grpc-client -addr=127.0.0.1:9090
# ingested: accepted=3 parsed=2 failed=1
# stats: total=2 counts=map[ERROR:1 INFO:1]
kill %1; rm -f /tmp/grpc-server /tmp/grpc-client

# Regenerate code from the .proto (requires protoc + plugins)
make proto
```

## Going further

### Read

- **gRPC-Go quick start + docs**: <https://grpc.io/docs/languages/go/> — the official tour.
- **Protocol Buffers language guide (proto3)**: <https://protobuf.dev/programming-guides/proto3/> — the type system + wire rules.
- **`google.golang.org/grpc/test/bufconn`** — how the in-process tests dial without a port.
- **`grpcurl`** — <https://github.com/fullstorydev/grpcurl> — curl for gRPC; pairs with server reflection.

### Try

- **Server-streaming a live tail** — add `rpc Tail(StatsRequest) returns (stream StatsReply)` that pushes a stats snapshot every second; client ranges over `stream.Recv()`.
- **Bidirectional streaming** — make `Ingest` bidi: ack each line as it's parsed (`rpc Ingest(stream LogLine) returns (stream LineAck)`).
- **Enable server reflection** — register `reflection.Register(srv)` and explore the service with `grpcurl -plaintext 127.0.0.1:9090 list`.
- **A timeout/retry `CallOption`** — give the client a per-call deadline (`context.WithTimeout`) and a retry interceptor; reuse L24's selective-retry thinking.
- **TLS** — swap `insecure.NewCredentials()` for `credentials.NewTLS(...)` with a self-signed cert; dial with the matching client creds.

---

> Phase 4 continues. Next: Lesson 26 — **Configuration, secrets & graceful shutdown**. The service learns flags > env > file config precedence, keeps secrets out of logs, and drains in-flight HTTP requests + gRPC calls cleanly on SIGTERM (`http.Server.Shutdown`, `grpc.GracefulStop`).
