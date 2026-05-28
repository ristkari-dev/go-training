<div class="title-slide-grid">
<div class="lesson-badge">
<div class="lesson-number">21</div>
<div class="lesson-word">Lesson</div>
</div>
<div class="title-slide-body">
<div class="lesson-label">Phase 3 — Concurrency &amp; Systems</div>
<h1>Networking &amp; syscalls</h1>
<div class="callout">
<h3>Learning goal</h3>
<p>Learn the <code>net</code> package: the TCP listen/accept loop, <code>net.Dial</code>, treating <code>net.Conn</code> as a streaming <code>io.ReadWriter</code>. Wire graceful shutdown via <code>signal.NotifyContext</code> + closing the listener. Understand the syscall layer underneath — every conn is an OS file descriptor, just like a file or pipe.</p>
</div>
</div>
</div>

---

## What we'll cover

- **TCP server** — `net.Listen` + accept loop + per-connection goroutine.
- **TCP client** — `net.Dial` + treating `net.Conn` as `io.ReadWriter`; half-close via `CloseWrite`.
- **Signals + graceful shutdown** — `signal.NotifyContext` cancels ctx; closing the listener unblocks `Accept`.
- **Syscalls + file descriptors** — every conn, file, and pipe is an OS FD; the `syscall` package; `ulimit -n`.

---

## The story so far

L16-L20 built up the **concurrency primitives**: goroutines, channels, select, sync, context, worker pools. Every example so far talked to itself — goroutines in one process passing data through in-memory channels.

**L21 reaches outside the process.** We use the same primitives — goroutines, ctx, defer — but now they're juggling real OS resources: sockets, file descriptors, signal handlers. The aggregator stays where L20 left it; this lesson is standalone. We build a tiny TCP echo server + client to ground the ideas.

---

## Concept 1: TCP server

### Motivation

You want to accept network connections. The standard library's `net` package gives you exactly what you need — `net.Listen` opens a server socket; `Accept` blocks until a client connects; then it's just `io.Reader`/`io.Writer` from there. Nothing fancy; no framework.

The shape is the same in every Go server you'll ever read: **listen → accept loop → per-connection goroutine**.

---

### The basics

```go
l, err := net.Listen("tcp", ":8080")
if err != nil { return err }
defer l.Close()

for {
    conn, err := l.Accept()
    if err != nil { return err }
    go handle(conn)
}
```

Three pieces:

- **`net.Listen("tcp", addr)`** returns a `net.Listener`. The addr is `host:port` — empty host means "all interfaces"; port `0` means "OS picks an ephemeral port."
- **`l.Accept()`** blocks until a client connects, then returns a `net.Conn` representing that connection.
- **`go handle(conn)`** spawns a goroutine per connection. This is the Go answer to C's thread-per-connection — except goroutines are cheap (a few KB each), so the same code that handles 10 connections handles 10,000.

`net.Conn` satisfies `io.Reader` + `io.Writer` + `io.Closer`. Anything you'd do with a file or buffer works on a TCP connection.

---

### A worked example

The L21 `echo-server` accept loop:

```go
for {
    conn, err := l.Accept()
    if err != nil {
        if ctx.Err() != nil {
            return nil  // graceful shutdown (see Concept 3)
        }
        return fmt.Errorf("accept: %w", err)
    }
    go func() {
        defer conn.Close()
        _ = lineecho.LineEcho(conn)
    }()
}
```

`lineecho.LineEcho(conn)` is just a `bufio.Scanner` over the conn, writing each line back uppercased via `bufio.Writer`. The conn is an `io.ReadWriter` — no networking code inside the handler.

`defer conn.Close()` in the goroutine is **non-negotiable** — each accepted conn holds an OS file descriptor; forgetting to close leaks FDs (see Concept 4).

---

### Common mistake

**Forgetting `defer conn.Close()` in the handler:**

```go
go func() {
    _ = handle(conn)  // ❌ never closes conn
}()
```

Every accepted conn opens an OS FD. Without `Close()`, the FD stays open until the process exits — even after the handler returns. A server that handles 10,000 connections in a day leaks 10,000 FDs, then crashes with "too many open files" once `ulimit -n` is hit (default 1024 on macOS, 1024-65536 on Linux).

The fix is the standard Go pattern — `defer` it on entry:

```go
go func() {
    defer conn.Close()
    _ = handle(conn)
}()
```

`go vet` doesn't catch this; it requires diligence. Make `defer conn.Close()` a reflex.

---

### Recap

- `net.Listen("tcp", addr)` → Accept loop → `go handle(conn)` is the universal server shape.
- `net.Conn` is an `io.ReadWriter` — handlers don't need network-specific code.
- **`defer conn.Close()`** is mandatory — every conn is an OS FD.

---

## Concept 2: TCP client

### Motivation

The mirror of the server side. `net.Dial` opens a connection to a server; the resulting `net.Conn` is again `io.ReadWriter`. The client looks structurally identical to the server's per-connection handler — same API on both ends.

What's different is **shape**: a client typically wants to read responses while still sending requests, which means a reader goroutine running concurrently with the main send loop.

---

### The basics

```go
conn, err := net.Dial("tcp", "localhost:8080")
if err != nil { return err }
defer conn.Close()

fmt.Fprintln(conn, "hello")      // write
r := bufio.NewScanner(conn)
r.Scan()                         // read
fmt.Println(r.Text())
```

Two things:

- **`net.Dial("tcp", addr)`** returns a `net.Conn`. The addr is `host:port` — DNS resolution happens here.
- **The conn IS the I/O.** Write to it, read from it. No "request" or "response" objects — that's a layer your code adds (HTTP, JSON-RPC, whatever).

For interactive clients (echo, REPL, SSH-like), the read and write sides need to run concurrently — otherwise you block on `Write` while a response sits unread.

---

### A worked example

The L21 `echo-client` runs the reader in its own goroutine + uses TCP **half-close** to drain cleanly on stdin EOF:

```go
conn, _ := net.Dial("tcp", addr)
defer conn.Close()

readerDone := make(chan struct{})
go func() {
    defer close(readerDone)
    r := bufio.NewScanner(conn)
    for r.Scan() {
        fmt.Fprintln(stdout, r.Text())
    }
}()

// Send stdin lines (buffered write, flush per line):
w := bufio.NewWriter(conn)
for stdin.Scan() {
    fmt.Fprintln(w, stdin.Text())
    w.Flush()
}

// Half-close: signal EOF on our send side; keep receive side open.
if tc, ok := conn.(*net.TCPConn); ok {
    _ = tc.CloseWrite()
}
<-readerDone  // wait for reader to drain remaining responses
```

`CloseWrite()` is the TCP half-close. It tells the peer "I'm done sending" via a FIN on the send side, while leaving the receive side open. The server's scanner sees EOF, finishes responding, then closes its end; our reader then sees its own EOF on the receive side and exits.

Without half-close, calling `conn.Close()` after the send loop would tear down BOTH directions immediately — and any responses still in-flight would be discarded.

---

### Common mistake

**No deadlines on long-lived conns:**

```go
conn, _ := net.Dial("tcp", "remote:8080")
defer conn.Close()
// ❌ no deadline; if the peer hangs, we hang forever
fmt.Fprintln(conn, "request")
r := bufio.NewScanner(conn)
r.Scan()  // blocks until response or TCP times out (minutes)
```

`net.Conn` has `SetDeadline(t time.Time)`, `SetReadDeadline`, and `SetWriteDeadline`. **Set them on any conn that talks to a remote you don't fully control.** A misbehaving peer (slow, malicious, broken NAT) can hang your client indefinitely otherwise.

```go
conn.SetReadDeadline(time.Now().Add(5 * time.Second))
r.Scan()  // now returns timeout error after 5s
```

For the echo-client demo we skip deadlines (localhost loopback is trustworthy). Production code talking over the public internet sets them on every read and every write.

---

### Recap

- `net.Dial("tcp", addr)` returns a `net.Conn`; same I/O API as the server side.
- Run the reader in a goroutine if you need concurrent read/write.
- **TCP half-close** (`CloseWrite`) signals "no more sends" without closing the receive side — lets responses drain.
- **Set deadlines** on conns talking to untrusted remotes.

---

## Concept 3: Signals + graceful shutdown

### Motivation

A server that runs forever should still STOP when asked. Pressing Ctrl-C, `kill -INT <pid>`, `docker stop`, Kubernetes pod termination — all deliver SIGINT or SIGTERM. A graceful shutdown does three things: stop accepting new connections, finish (or cleanly abandon) in-flight ones, and exit zero.

`signal.NotifyContext` is the modern Go answer: wrap your root ctx so OS signals cancel it. Combined with a closer goroutine that closes the listener on cancel, the accept loop tears down cleanly without raceful gymnastics.

---

### The basics

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

// ctx is now cancelled when SIGINT arrives.
```

`signal.NotifyContext(parent, signals...)` (Go 1.16+) returns a derived ctx that's cancelled when any of the named signals arrive. `defer stop()` removes the signal handler when main returns — same shape as `defer cancel()`.

To unblock an in-flight `Accept`, close the listener:

```go
go func() {
    <-ctx.Done()
    _ = l.Close()  // makes Accept return an error
}()

for {
    conn, err := l.Accept()
    if err != nil {
        if ctx.Err() != nil { return nil }  // graceful shutdown
        return err
    }
    // ...
}
```

The dance: ctx cancels → closer goroutine closes listener → next `Accept()` errors → loop checks `ctx.Err()` to distinguish "we closed it on purpose" from "real failure."

---

### A worked example

The L21 echo-server's full shutdown wiring:

```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
    if err := run(ctx, addr, os.Stdout); err != nil { /* ... */ }
}

func run(ctx context.Context, addr string, stdout io.Writer) error {
    l, _ := net.Listen("tcp", addr)
    defer l.Close()

    go func() {
        <-ctx.Done()
        _ = l.Close()
    }()

    for {
        conn, err := l.Accept()
        if err != nil {
            if ctx.Err() != nil { return nil }
            return fmt.Errorf("accept: %w", err)
        }
        go func() {
            defer conn.Close()
            _ = lineecho.LineEcho(conn)
        }()
    }
}
```

Two `l.Close()` calls — the deferred one and the closer goroutine's — are both fine; `Close` is idempotent.

In-flight connection handlers may still be running when `run` returns. For short-lived work (line echo) that's acceptable. For long-lived handlers (long polling, file uploads), wrap each handler goroutine in a `sync.WaitGroup` and wait for them before returning.

---

### Common mistake

**Not handling the "graceful close vs real failure" distinction in the accept loop:**

```go
for {
    conn, err := l.Accept()
    if err != nil {
        return fmt.Errorf("accept: %w", err)  // ❌ always treats as error
    }
    go handle(conn)
}
```

When you close the listener on purpose, `Accept` returns an error too — but it's not a real failure. The server logs "accept: use of closed network connection" and exits 1. Indistinguishable from a real listener crash.

Fix: check `ctx.Err()` after the accept error:

```go
if err != nil {
    if ctx.Err() != nil { return nil }  // we closed it
    return fmt.Errorf("accept: %w", err)
}
```

The same pattern — "did we cancel ctx, or did something else fail?" — shows up everywhere in Go server code.

---

### Recap

- `signal.NotifyContext(ctx, os.Interrupt)` cancels ctx on Ctrl-C.
- Close the listener from a closer goroutine on `<-ctx.Done()` — unblocks `Accept`.
- After accept error, check `ctx.Err()` to distinguish graceful shutdown from real failure.

---

## Concept 4: Syscalls + file descriptors

### Motivation

`net.Conn`, `*os.File`, pipes, terminals — at the OS level, all of them are **file descriptors**. A small integer the kernel maps to a kernel object (a socket, a vnode, a pipe end). Go hides the integer behind types, but the constraint is shared: every open FD costs you one, and there's a hard ceiling.

Knowing this lets you read kernel error messages, debug FD leaks, and understand why every example in this lesson `defer`s `Close()`.

---

### The basics

`ulimit -n` shows your process's FD ceiling:

```bash
$ ulimit -n
1024              # default on most macOS / many Linux distros
```

Every open file, every open conn, every pipe end consumes one slot. Hit the ceiling and the next `Open`/`Accept`/`Dial` returns `EMFILE` ("too many open files"):

```
accept tcp [::]:8080: accept4: too many open files
```

You can usually raise the ceiling — `ulimit -n 65536` for the shell, or per-process via `setrlimit`. But raising it is treating the symptom; the real fix is closing what you open.

Helpful tools:

- **`lsof -p <pid>`** — list every FD the process has open. Run on a suspected leaker to see what's accumulating.
- **`/proc/<pid>/fd/`** (Linux) — directory of symlinks, one per FD.
- **Activity Monitor → Open Files & Ports** (macOS) — GUI equivalent.

---

### A worked example

The Go `syscall` package exposes raw kernel calls. `net.Conn`'s `SyscallConn().Control(func(fd uintptr))` lets you reach the underlying FD:

```go
tc := conn.(*net.TCPConn)
raw, _ := tc.SyscallConn()
raw.Control(func(fd uintptr) {
    fmt.Println("conn FD =", fd)  // a small integer, e.g. 7
})
```

In production code, you almost never need this. It exists for the rare moments where you want to tune socket options the stdlib doesn't expose (e.g., `SO_REUSEPORT` for load-balanced listeners), or interoperate with C libraries that take an FD.

Three things that go through the FD abstraction:

| Type             | What's the FD pointing to       |
|------------------|----------------------------------|
| `*os.File`       | A file, directory, or device     |
| `net.Conn`       | A socket (TCP, UDP, Unix)        |
| Pipe (`os.Pipe`) | A pipe end (read or write side)  |

All three implement `io.Reader`/`io.Writer` — that's the whole point of the abstraction. Code that works on `io.Reader` works whether you handed it a file, a socket, or a pipe.

---

### Common mistake

**FD leaks via missing `Close()`:**

```go
files, _ := os.ReadDir("/tmp")
for _, f := range files {
    file, _ := os.Open(filepath.Join("/tmp", f.Name()))
    // ❌ never closed; the loop opens N FDs and only returns them on process exit
    _ = process(file)
}
```

Each iteration opens one FD. By iteration 1024, the program crashes with `EMFILE`. The fix is `defer file.Close()` inside the loop body — but `defer` inside a loop fires at function return, not loop iteration. The cleaner shape:

```go
for _, f := range files {
    err := func() error {
        file, err := os.Open(filepath.Join("/tmp", f.Name()))
        if err != nil { return err }
        defer file.Close()      // ✓ fires at end of THIS iteration
        return process(file)
    }()
    if err != nil { return err }
}
```

The same pattern applies to conns: never let a `defer` for a per-connection resource live in the accepting function — push it into the handler goroutine.

---

### Recap

- Every conn, file, and pipe is an OS file descriptor (FD).
- `ulimit -n` caps how many your process can hold; `EMFILE` is what you hit at the ceiling.
- `lsof -p <pid>` shows what's open.
- **FD leaks come from missing `Close()`** — `defer` it at the same scope as the open.

---

## Practice

### Warm-up

Implement `LineEcho(conn net.Conn) error` in `exercises/warmup/lineecho/lineecho.go`. Read lines via `bufio.Scanner`; write each back uppercased via `bufio.Writer` (flush per line). Tests use `net.Pipe()` for in-memory connection pairs — no real TCP needed.

```bash
cd lessons/21-networking/exercises/warmup/lineecho
go test -v
go test -race
```

---

### Main

Two CLI binaries, each ~50 lines:

1. **`cmd/echo-server`** — `net.Listen` + accept loop + per-connection `lineecho.LineEcho`. Graceful shutdown via `signal.NotifyContext` + closer goroutine. `-addr=:0` (default) lets the OS pick an ephemeral port; the chosen addr is logged so tests can parse it.

2. **`cmd/echo-client`** — `net.Dial` + reader goroutine for responses + main loop reading stdin and sending to server. Half-close (`CloseWrite`) on stdin EOF + drain via `<-readerDone` before returning.

```bash
cd lessons/21-networking/exercises
go test ./...
go test -race ./...
make test-race      # daily habit (since L18)
```

Manual smoke test (in two terminals):

```bash
# Terminal 1:
go run ./lessons/21-networking/solutions/cmd/echo-server -addr=:9876
# Output: listening on [::]:9876

# Terminal 2:
echo "hello world" | nc localhost 9876
# Output: HELLO WORLD

# Or use our client:
echo "hello world" | go run ./lessons/21-networking/solutions/cmd/echo-client -addr=localhost:9876
# Output: HELLO WORLD
```

---

## Closing thought

The `net` package is small. Five functions and an interface (`Listen`, `Dial`, `Pipe`, `Conn`, `Listener`) cover 90% of networking code you'll ever write. Everything else — HTTP, gRPC, WebSockets, SSH — is built on top.

That smallness is by design. Once you know `net.Conn` is just an `io.ReadWriter` with `Close` and a couple of deadline knobs, every higher-level package becomes "oh, that's just a protocol on top of bytes." HTTP is "lines in, lines out, then a body." TLS is "wrap a conn in another conn that encrypts." gRPC is "framed messages over HTTP/2."

L22 (the last Phase 3 lesson) takes this server and **measures it** — pprof CPU profiles, memory allocations, benchmarks for the aggregator's three Walk variants. Once you can build it and observe it, you can make it fast.

---

## What we learned

- **TCP server**: `net.Listen` + accept loop + per-connection goroutine. `defer conn.Close()` is mandatory.
- **TCP client**: `net.Dial` + treat `net.Conn` as `io.ReadWriter`. Use a reader goroutine for concurrent send/receive; half-close to drain cleanly.
- **Graceful shutdown**: `signal.NotifyContext` cancels ctx; closer goroutine closes the listener; accept loop checks `ctx.Err()` to distinguish graceful from real failure.
- **FDs**: conns, files, and pipes all share the OS file-descriptor abstraction. `ulimit -n` caps you; missing `Close()` leaks; `lsof` shows what's open.

---

## Up next

Lesson 22 — **Profiling & benchmarks**. The Phase 3 finale. We bring `pprof` to the aggregator: CPU profile, heap profile, allocation traces. We add `go test -bench` to the three Walk variants from L20 and find out which one wins on YOUR machine, and why. Then Phase 4 begins — HTTP, gRPC, observability, distributed systems.
