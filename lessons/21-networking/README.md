# Lesson 21: Networking & syscalls

## What you'll learn

By the end of this lesson you can:

- Open a TCP server with **`net.Listen` + accept loop + per-connection goroutine** — the universal Go server shape.
- Open a TCP client with **`net.Dial`** and treat the resulting `net.Conn` as an `io.ReadWriter`. Use a reader goroutine for concurrent send/receive; **half-close** (`CloseWrite`) to drain responses cleanly.
- Wire **graceful shutdown** via `signal.NotifyContext` + a closer goroutine that closes the listener on `<-ctx.Done()`.
- Reason about **file descriptors** — every conn, file, and pipe is an OS FD; the `ulimit -n` ceiling; FD-leak debugging.
- Build a tiny TCP echo server + client end to end, with in-process integration tests that need no real network.

## What's different from L20

L20 was the Phase 3 concurrency-patterns capstone — three Walk variants in the aggregator (channels, mutex, worker pool). **L21 leaves the aggregator alone** and pivots to networking. We use the same primitives (goroutines, ctx, defer, select) but they're now juggling real OS resources: sockets, file descriptors, signal handlers.

The lesson is standalone: no carry-forward of the L14-L20 logparse/aggregator code. Two small CLI binaries (`echo-server`, `echo-client`) plus a one-function warmup (`lineecho.LineEcho`) cover the four concepts.

## The package layout

```
lessons/21-networking/exercises/
├── warmup/lineecho/                ← you implement: LineEcho(conn net.Conn) error
│   ├── lineecho.go
│   └── lineecho_test.go            (skeleton; uses net.Pipe() in solutions)
└── cmd/
    ├── echo-server/                ← you implement: TCP server + graceful shutdown
    │   ├── main.go
    │   └── main_test.go            (skeleton; in-process integration test in solutions)
    └── echo-client/                ← you implement: TCP client + half-close drain
        ├── main.go
        └── main_test.go            (skeleton; in-process integration test in solutions)
```

---

## Concept 1 — TCP server

`net.Listen("tcp", addr)` opens a server socket; `Accept()` blocks until a client connects; spawn a goroutine per connection. `net.Conn` satisfies `io.ReadWriter`, so handlers don't need network-specific code — just `bufio.Scanner` or `io.Copy` like you'd use on a file.

```go
l, err := net.Listen("tcp", ":8080")
if err != nil { return err }
defer l.Close()

for {
    conn, err := l.Accept()
    if err != nil { return err }
    go func() {
        defer conn.Close()
        handle(conn)
    }()
}
```

Three things:

- **`net.Listen`** binds a port. Empty host = all interfaces; port `0` = ephemeral (OS picks).
- **`Accept()`** blocks until a client connects.
- **`go handle(conn)`** is the Go answer to thread-per-connection. Goroutines cost a few KB each, so the same code that handles 10 conns handles 10,000.

### Common mistake

**Forgetting `defer conn.Close()`** in the handler goroutine. Every accepted conn opens an OS file descriptor. Without `Close`, the FD stays open until the process exits — even after the handler returns. A long-running server leaks FDs until it hits `ulimit -n` (default 1024 on macOS) and crashes with "too many open files."

The fix is reflex: `defer conn.Close()` is the first line of every handler goroutine.

---

## Concept 2 — TCP client

`net.Dial("tcp", "host:port")` opens a connection; you get back a `net.Conn` — same `io.ReadWriter` type as on the server side. For interactive clients (REPL, echo, SSH-like), run the reader in its own goroutine so responses appear as they arrive instead of blocking the send loop.

```go
conn, _ := net.Dial("tcp", "localhost:8080")
defer conn.Close()

readerDone := make(chan struct{})
go func() {
    defer close(readerDone)
    r := bufio.NewScanner(conn)
    for r.Scan() { fmt.Println(r.Text()) }
}()

for stdin.Scan() {
    fmt.Fprintln(conn, stdin.Text())
}

// Half-close: send-side FIN, keep receive side open.
if tc, ok := conn.(*net.TCPConn); ok { _ = tc.CloseWrite() }
<-readerDone   // drain remaining responses before exit
```

**TCP half-close** (`CloseWrite`) is the trick that makes this drain cleanly. After stdin EOF, calling plain `conn.Close()` would tear down BOTH directions immediately — and any responses still in-flight (already-buffered by the kernel, not yet read by our reader goroutine) get discarded. Half-close signals "I'm done sending" on our send side, the server sees EOF, finishes responding, then closes its end; our reader sees EOF and exits.

### Common mistake

**No deadlines on long-lived conns.** `net.Conn` has `SetDeadline`, `SetReadDeadline`, and `SetWriteDeadline`. A misbehaving peer (slow, malicious, broken NAT) can hang your client indefinitely otherwise — the kernel's TCP keepalive timer is minutes-long. For any conn talking to a remote you don't fully control, set deadlines on every read and every write.

```go
conn.SetReadDeadline(time.Now().Add(5 * time.Second))
r.Scan()  // now returns timeout error after 5s if no data arrives
```

The L21 echo-client skips deadlines because it's localhost-only; production code does not have that luxury.

---

## Concept 3 — Signals + graceful shutdown

A server that runs forever should still STOP cleanly when asked. Ctrl-C / `kill -INT` / `docker stop` / k8s pod termination all send SIGINT or SIGTERM. `signal.NotifyContext` (Go 1.16+) is the modern wire-up:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()
```

`ctx` is now cancelled when SIGINT arrives. To unblock an in-flight `Accept`, close the listener from a closer goroutine:

```go
go func() {
    <-ctx.Done()
    _ = l.Close()  // makes Accept return an error
}()

for {
    conn, err := l.Accept()
    if err != nil {
        if ctx.Err() != nil { return nil }    // graceful: we closed it
        return fmt.Errorf("accept: %w", err)  // real failure
    }
    go handle(conn)
}
```

The choreography: ctx cancels → closer closes listener → next `Accept()` errors → loop checks `ctx.Err()` to tell graceful shutdown from real failure → exit cleanly.

In-flight handlers may still be running when `run` returns. For short work (echo, lookups) this is acceptable. For long handlers (uploads, long polling) wrap each handler goroutine in a `sync.WaitGroup` and `wg.Wait()` before returning — same pattern as L18's WalkLocked.

### Common mistake

**Not distinguishing "we closed it" from "it crashed"** in the accept loop. Without the `ctx.Err() != nil → return nil` check, every graceful shutdown prints "accept: use of closed network connection" and exits 1 — indistinguishable from a real listener crash. CI sees the non-zero exit and fails the deployment. The two-line check fixes it.

---

## Concept 4 — Syscalls + file descriptors

`net.Conn`, `*os.File`, pipes, terminals — at the OS level, all of them are **file descriptors**. A small integer the kernel maps to a kernel object (socket, vnode, pipe end). Go hides the integer behind types, but the constraint is shared: every open FD costs you one, and `ulimit -n` caps how many you can hold.

```bash
$ ulimit -n
1024              # default on most macOS / many Linux distros
```

Hit the ceiling and the next `Open`/`Accept`/`Dial` returns `EMFILE` ("too many open files"). You can raise it (`ulimit -n 65536` for a shell), but the real fix is closing what you open.

Useful tools:

- **`lsof -p <pid>`** — list every FD the process holds. Run on a suspected leaker to see what's accumulating.
- **`/proc/<pid>/fd/`** (Linux only) — directory of symlinks, one per FD.
- **`syscall` package** — raw kernel calls. You rarely need this directly; `net.Conn.SyscallConn().Control(func(fd uintptr))` is the escape hatch when you do (e.g., `SO_REUSEPORT`, custom socket options).

Three types share the FD abstraction:

| Go type          | FD points to                     |
|------------------|----------------------------------|
| `*os.File`       | File, directory, device          |
| `net.Conn`       | Socket (TCP, UDP, Unix)          |
| Pipe (`os.Pipe`) | Pipe end (read or write side)    |

All three implement `io.Reader`/`io.Writer`. Code that takes `io.Reader` works whether you handed it a file, a socket, or a pipe — that's the whole point of the abstraction.

### Common mistake

**FD leaks via missing `Close()`**, especially inside loops where naive `defer file.Close()` is wrong (it fires at function return, not loop iteration). The cleaner shape wraps the per-iteration work in an inner function so `defer` fires per-iteration:

```go
for _, f := range files {
    err := func() error {
        file, err := os.Open(filepath.Join("/tmp", f.Name()))
        if err != nil { return err }
        defer file.Close()         // ✓ fires at end of THIS iteration
        return process(file)
    }()
    if err != nil { return err }
}
```

The same pattern applies to accepted conns: the `defer conn.Close()` MUST live in the handler goroutine, not the accept loop.

---

## Exercise: warm-up — `lineecho`

Implement `LineEcho(conn net.Conn) error` in `exercises/warmup/lineecho/lineecho.go`. Read lines via `bufio.Scanner`; write each back uppercased via `bufio.Writer` (flush per line). Tests use `net.Pipe()` — an in-memory pair of `net.Conn` values — so the test never opens a real socket.

**Time:** 10-15 minutes.

## Exercise: main — `cmd/echo-server` + `cmd/echo-client`

Implement both binaries:

1. **`cmd/echo-server/main.go`** — `net.Listen` on `-addr` (default `:0` for ephemeral); accept loop; per-connection goroutine calls `lineecho.LineEcho`; graceful shutdown via `signal.NotifyContext` + closer goroutine.

2. **`cmd/echo-client/main.go`** — `net.Dial` on `-addr` (required, no default); reader goroutine prints server responses; main loop reads stdin and sends to server; half-close + drain on stdin EOF.

Tests are in-process integration tests: spawn the server's `run(ctx, ":0", &log)` in a goroutine, parse the chosen port from the captured log, dial it, send/recv, cancel ctx. No real port is bound to a fixed number; everything is ephemeral.

**Time:** 45-60 minutes.

---

## Daily habits

After every change:

```bash
gofmt -w ./...
go vet ./...
go test ./...
make test-race    # daily habit since L18
```

## How to run

```bash
# Skeleton (passes vacuously until you implement)
cd lessons/21-networking/exercises
go test ./...

# Reference solution
go test -v ./lessons/21-networking/solutions/...

# Race detector
go test -race ./lessons/21-networking/solutions/...
make test-race

# Manual smoke (two terminals)
# Terminal 1:
go run ./lessons/21-networking/solutions/cmd/echo-server -addr=:9876
# Output: listening on [::]:9876

# Terminal 2 (with our client):
echo "hello world" | go run ./lessons/21-networking/solutions/cmd/echo-client -addr=localhost:9876
# Output: HELLO WORLD

# Or with netcat:
echo "hello world" | nc localhost 9876
# Output: HELLO WORLD
```

## Going further

### Read

- **Go std-lib — `net` package**: <https://pkg.go.dev/net> — surprisingly readable; skim once and the whole package becomes navigable.
- **Go std-lib — `os/signal`**: <https://pkg.go.dev/os/signal> — covers `Notify`, `NotifyContext`, signal-ignore patterns.
- **"The Go Programming Language"** (Donovan & Kernighan), Chapter 8 — concurrent server examples in the same shape we used here.
- **Beej's Guide to Network Programming** — <https://beej.us/guide/bgnet/> — language-agnostic intro to socket programming. If you only know Go networking, this is the C-level mental model worth borrowing.

### Try

- **Add a `-timeout` flag to echo-server**: each handler gets a per-connection read deadline (`conn.SetReadDeadline`). After N seconds of silence, the handler closes the conn. Verify with a hanging `nc localhost 9876` that doesn't send anything.
- **Wait for in-flight handlers on shutdown**: extend echo-server's `run` to track handler goroutines via `sync.WaitGroup` and `wg.Wait()` before returning. Verify: connect a slow client (one that types one line then sleeps 30s), then SIGINT the server; server should wait for the slow client to finish.
- **HTTP server from scratch**: rewrite echo-server as a tiny HTTP server (no `net/http`). Parse the request line manually (`GET / HTTP/1.1\r\n`), parse headers, write `HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\nHello, world`. Compare to `net/http`'s implementation. ~50 lines of fun.
- **Unix domain socket variant**: change `net.Listen("tcp", ...)` to `net.Listen("unix", "/tmp/echo.sock")`. The rest of the code is identical — same `net.Conn`, same handler. Useful for inter-process communication on the same host.

---

> Phase 3 is almost done. Next stop: Lesson 22 — Profiling & benchmarks. We bring `pprof` to the aggregator and benchmark the three Walk variants from L20 to find the saturation point on your machine. Phase 4 (HTTP, gRPC, observability, distributed systems) starts after that.
