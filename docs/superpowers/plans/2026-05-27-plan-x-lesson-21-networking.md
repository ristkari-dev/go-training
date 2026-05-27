# Plan X — Lesson 21 (Networking & syscalls) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Author lesson 21 of Phase 3 — standalone networking lesson. Students learn the `net` package, TCP listen/accept loop, `net.Dial`, signal-based graceful shutdown, and the syscall/FD layer beneath. Build a tiny TCP echo server + client. Per Phase 3 spec, the aggregator stays where L20 left it — no carry-forward.

**Architecture:** Same per-lesson pattern as Plans D-W. Six tasks. Skeleton tests in warmup + both cmd binaries (Phase 3 invariant). Heavy-explanatory slide deck with **four** concept blocks (TCP server, TCP client, signals + graceful shutdown, syscalls + FDs).

**Tech Stack:** Go 1.23 stdlib only (`bufio`, `context`, `errors`, `flag`, `fmt`, `io`, `net`, `os`, `os/signal`, `strings`, `sync`, `testing`, `time`). Reveal.js 5.1.0.

---

## Scope

After Plan X: lesson 21 complete; `make test` + `make test-race` green; both binaries work end-to-end; in-process integration tests pass.

### Design decisions (2 user-approved + 7 plan-recommended)

**User-approved via brainstorming:**

1. **Four slide concepts**: TCP server, TCP client, signals + graceful shutdown, syscalls + FDs. Matches L19's 4-concept rhythm. Connection lifecycle gotchas (deadlines, half-close) deferred to L22 or Phase 4.

2. **All logic in `cmd/echo-server/main.go` + `cmd/echo-client/main.go`** with testable `run(ctx, ...)` entry points. No internal package extraction — matches L13/L17/L19 cmd/aggregator precedent. In-process integration via `net.Listen("tcp", "127.0.0.1:0")`.

**Plan-recommended:**

3. **Warmup `lineecho.LineEcho(conn net.Conn) error`**: read lines via `bufio.Scanner`; write uppercased back via `bufio.Writer` (flush per line). Tests use `net.Pipe()` for in-memory testing (no real network involved).

4. **Server: `-addr` flag defaults to `:0` (ephemeral port)**. Server logs `listening on <actual addr>` after binding. CLI usage: `-addr=:8080` for fixed port; `-addr=:0` to let OS pick. The ephemeral-port logging is essential for the test (parses the log line to get the port).

5. **Graceful shutdown**: server's `main` uses `signal.NotifyContext(ctx, os.Interrupt)`; on ctx cancellation, closes the listener (which makes the accept loop's `Accept()` return an error); accept loop exits; main returns clean. In-flight connection handlers may still be running — for the lesson's scope, that's acceptable (handlers are short-lived line-echo work).

6. **Client: stdin-piped**: reads lines from stdin via `bufio.Scanner`; sends each + "\n" to server; reads response line from server; writes to stdout. Exits on stdin EOF (Ctrl-D) or ctx cancellation (SIGINT).

7. **In-process integration tests** spawn server's `run(ctx, ":0", ...)` in a goroutine; parse the logged port from the captured writer; dial it; assert response; cancel ctx. Same test pattern as L19's TestAggregatorCancelled — proven.

8. **No `internal/` package**: per Q2 decision. All code in cmd binaries. Each cmd's `main_test.go` has its own integration test using the in-process pattern.

9. **`net.Pipe()` for warmup tests**: avoids real TCP setup for unit testing LineEcho. `net.Pipe()` returns two `net.Conn` halves wired in-memory — perfect for testing functions that take `net.Conn` without needing a listener.

---

## Plans F-W lessons-learned applied here

1. Subpackages have their own namespace — no `Warmup*` prefix.
2. Lint exclusion covers nested subpackages.
3. Common-mistake content in README (4 per lesson, one per concept).
4. Slides + README written inline by controller.
5. `make test-race` daily-habits continues from L18.
6. Skeleton tests pass vacuously.
7. Scaffold + restructure (delete 8 flat scaffolder files).
8. gofmt 1.19+ godoc list normalization tolerance.
9. **CI flake lessons from L17-L20**: avoid content-specific assertions that depend on timing. Network tests should assert on data flow shape (e.g., "stdout contains the uppercased response"), not exact byte counts that could be affected by buffering.
10. **L21 slug is already correct** (`networking`); no build-index fix needed.

---

## File Structure

After Plan X (~14 files — smaller than aggregator lessons):

```
lessons/21-networking/
├── README.md                                       (Task 6)
├── slides/{index.html, slides.md, assets/.gitkeep} (Task 1 + Task 5)
├── exercises/
│   ├── warmup/lineecho/
│   │   ├── lineecho.go                             (Task 2)
│   │   └── lineecho_test.go                        (Task 2 — SKELETON)
│   ├── cmd/echo-server/
│   │   ├── main.go                                 (Task 3)
│   │   └── main_test.go                            (Task 3 — SKELETON)
│   └── cmd/echo-client/
│       ├── main.go                                 (Task 4)
│       └── main_test.go                            (Task 4 — SKELETON)
└── solutions/  (mirrored)
```

---

## Conventions

- **Branch:** `feature/plan-x-lesson-21-networking`
- **Commit messages:** Conventional Commits
- **No carry-forward sources** (lesson is standalone)

---

## Task 1: Scaffold + restructure

Same dance as Plans J-W. No slug fix needed (slug `networking` matches directory).

- [ ] **Step 1:** `make new-lesson NAME=21-networking`
- [ ] **Step 2:** Delete 8 flat scaffolder files.
- [ ] **Step 3:** Verify 4-file scaffolded tree + slides build + dist/lessons/21-networking/ created.
- [ ] **Step 4:** Commit:

```bash
git add lessons/21-networking/
git commit -m "feat(lessons): scaffold lesson 21-networking with empty subpackage layout"
```

---

## Task 2: Author the warm-up — `lineecho`

`LineEcho(conn net.Conn) error` — read lines, write uppercased back. Tests use `net.Pipe()` for in-memory testing.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/21-networking/{exercises,solutions}/warmup/lineecho`

- [ ] **Step 2:** Create `lessons/21-networking/exercises/warmup/lineecho/lineecho.go`:

```go
// Package lineecho is the lesson 21 warm-up: read lines from a net.Conn,
// write each back uppercased.
//
// LineEcho is the per-connection handler the echo-server uses. The
// caller is responsible for opening and closing the conn; LineEcho
// just reads lines until EOF and writes uppercased responses.
//
// Tested with net.Pipe() — an in-memory pair of net.Conn values. The
// pipe lets us test LineEcho without setting up a real TCP listener,
// because net.Conn is an interface and net.Pipe() returns conforming
// in-memory implementations.
package lineecho

import (
	"bufio"
	"net"
	"strings"
)

// LineEcho reads lines from conn via bufio.Scanner, writes each back
// uppercased via bufio.Writer (flushed per line). Returns nil on
// clean EOF (peer closed). Returns the scanner's error if reads fail
// mid-stream.
//
// Does NOT close conn — that's the caller's responsibility.
//
// Hint:
//   1. r := bufio.NewScanner(conn)
//   2. w := bufio.NewWriter(conn)
//   3. for r.Scan() {
//        upper := strings.ToUpper(r.Text())
//        if _, err := fmt.Fprintln(w, upper); err != nil { return err }
//        if err := w.Flush(); err != nil { return err }    // flush per line
//      }
//   4. return r.Err()    // nil on clean EOF
//
// Flush per line is essential for an interactive protocol — the client
// is waiting for a response after each line, so the response must
// actually be sent (not sit in the buffer).
func LineEcho(conn net.Conn) error {
	_ = bufio.NewScanner
	_ = strings.ToUpper
	panic("TODO: bufio.Scanner read loop + bufio.Writer per-line flush; return scanner.Err()")
}
```

- [ ] **Step 3:** Create `lessons/21-networking/exercises/warmup/lineecho/lineecho_test.go` (SKELETON):

```go
package lineecho

import (
	"bufio"
	"net"
	"testing"
)

// TestLineEcho is a SKELETON. Use net.Pipe() for in-memory testing.
// Spawn LineEcho on one end in a goroutine; write lines on the other;
// read back uppercased responses.
func TestLineEcho(t *testing.T) {
	cases := []struct {
		name string
		send string
		want string
	}{
		// TODO: at least 3 cases.
		// {"single", "hello\n", "HELLO\n"},
		// {"multiple", "hi\nworld\n", "HI\nWORLD\n"},
		// {"mixed-case", "Hello World\n", "HELLO WORLD\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO:
			//   serverEnd, clientEnd := net.Pipe()
			//   defer serverEnd.Close(); defer clientEnd.Close()
			//   go func() { _ = LineEcho(serverEnd); serverEnd.Close() }()
			//   client writes tc.send and closes its write side (or just sends)
			//   client reads response via bufio.Scanner
			//   collect lines; compare to tc.want
			_ = tc
			_ = bufio.NewScanner
			_ = net.Pipe
		})
	}
}

// TestLineEchoEmpty is a SKELETON. Empty input (peer closes immediately)
// → returns nil.
func TestLineEchoEmpty(t *testing.T) {
	// TODO: net.Pipe(), close the client side immediately, call LineEcho on
	// server side, assert nil error.
}
```

- [ ] **Step 4:** Create `lessons/21-networking/solutions/warmup/lineecho/lineecho.go`:

```go
// Package lineecho is the lesson 21 warm-up reference implementation.
package lineecho

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// LineEcho reads lines, writes each back uppercased.
func LineEcho(conn net.Conn) error {
	r := bufio.NewScanner(conn)
	w := bufio.NewWriter(conn)

	for r.Scan() {
		upper := strings.ToUpper(r.Text())
		if _, err := fmt.Fprintln(w, upper); err != nil {
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return r.Err()
}
```

- [ ] **Step 5:** Create `lessons/21-networking/solutions/warmup/lineecho/lineecho_test.go`:

```go
package lineecho

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestLineEcho(t *testing.T) {
	cases := []struct {
		name string
		send []string
		want []string
	}{
		{"single", []string{"hello"}, []string{"HELLO"}},
		{"multiple", []string{"hi", "world"}, []string{"HI", "WORLD"}},
		{"mixed-case", []string{"Hello World"}, []string{"HELLO WORLD"}},
		{"unicode", []string{"héllo"}, []string{"HÉLLO"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			serverEnd, clientEnd := net.Pipe()

			// Run LineEcho on the server side; signal completion via done channel.
			done := make(chan error, 1)
			go func() {
				done <- LineEcho(serverEnd)
				_ = serverEnd.Close()
			}()

			// Client writes each line + collects responses.
			go func() {
				w := bufio.NewWriter(clientEnd)
				for _, line := range tc.send {
					fmt.Fprintln(w, line)
				}
				w.Flush()
				// Close the client-write side to signal EOF; but net.Pipe()
				// doesn't have a separate write side, so we have to close
				// the whole connection.
				// To avoid prematurely closing the read side, we wait until
				// we've read all responses below — done in main test goroutine.
			}()

			// Read responses from client side.
			r := bufio.NewScanner(clientEnd)
			var got []string
			for i := 0; i < len(tc.want); i++ {
				if !r.Scan() {
					t.Fatalf("expected %d responses, got %d (err: %v)", len(tc.want), i, r.Err())
				}
				got = append(got, r.Text())
			}

			// Close client side to signal EOF to LineEcho.
			_ = clientEnd.Close()

			// Wait for LineEcho to finish.
			if err := <-done; err != nil {
				t.Errorf("LineEcho returned error: %v", err)
			}

			if !equalStrings(got, tc.want) {
				t.Errorf("LineEcho got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLineEchoEmpty(t *testing.T) {
	serverEnd, clientEnd := net.Pipe()

	done := make(chan error, 1)
	go func() {
		done <- LineEcho(serverEnd)
		_ = serverEnd.Close()
	}()

	// Close client side immediately — LineEcho sees EOF.
	_ = clientEnd.Close()

	if err := <-done; err != nil {
		t.Errorf("LineEcho with empty input returned error: %v", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// (using strings to keep gofmt happy with the import; remove if not needed)
var _ = strings.ToUpper
```

> Note on `net.Pipe()`: returns two `net.Conn` halves connected in-memory. Writes on one show up as reads on the other. Perfect for testing functions that take `net.Conn` without needing a real listener.

> Note on the test's close-order dance: net.Pipe doesn't have separate read/write halves, so we can't "close the write side" — we have to close the whole conn. The test reads all expected responses FIRST, then closes the client side, which signals EOF to LineEcho's scanner.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/21-networking/
go test ./lessons/21-networking/exercises/warmup/lineecho/... -v 2>&1 | tail -10
go test ./lessons/21-networking/solutions/warmup/lineecho/... -v 2>&1 | tail -20
go test -race ./lessons/21-networking/solutions/warmup/lineecho/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/21-networking/exercises/warmup/ lessons/21-networking/solutions/warmup/
git commit -m "feat(lesson-21): warmup — lineecho (read+uppercase+write via bufio over net.Conn)"
```

Expected: gofmt empty; exercises pass vacuously; solutions tests PASS; -race clean.

---

## Task 3: cmd/echo-server

Listen on `-addr` (default `:0`), accept loop, per-connection goroutine calls LineEcho. Graceful shutdown via signal.NotifyContext.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/21-networking/{exercises,solutions}/cmd/echo-server`

- [ ] **Step 2:** Create `lessons/21-networking/exercises/cmd/echo-server/main.go`:

```go
// Package main is the lesson 21 TCP echo server.
//
// Listens on -addr (default :0 for ephemeral port); accept loop spawns
// one goroutine per connection; each goroutine calls lineecho.LineEcho.
//
// Graceful shutdown: SIGINT cancels ctx; main closes the listener,
// which makes Accept return an error; accept loop exits cleanly.
// In-flight handlers may still be running — for this lesson's scope
// (short-lived line-echo work) that's acceptable.
//
// Usage:
//   echo-server -addr=:8080    # fixed port
//   echo-server -addr=:0       # ephemeral port (default; logged after bind)
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/21-networking/exercises/warmup/lineecho"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := parseAddr(os.Args[1:])

	if err := run(ctx, addr, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point.
//
// Hint:
//   1. l, err := net.Listen("tcp", addr)
//   2. if err != nil → return err
//   3. defer l.Close()
//   4. fmt.Fprintf(w, "listening on %s\n", l.Addr())
//   5. Closer goroutine: <-ctx.Done(); l.Close()
//   6. Accept loop:
//        for {
//            conn, err := l.Accept()
//            if err != nil {
//                if ctx.Err() != nil { return nil }  // graceful shutdown
//                return err
//            }
//            go func() {
//                defer conn.Close()
//                _ = lineecho.LineEcho(conn)
//            }()
//        }
//
// The "ctx.Err() != nil → return nil" pattern is the canonical
// graceful-shutdown signal: if the listener errored AFTER ctx was
// cancelled, it's because we closed it on purpose, not a real failure.
func run(ctx context.Context, addr string, stdout io.Writer) error {
	_ = net.Listen
	_ = lineecho.LineEcho
	panic("TODO: net.Listen, log actual addr, closer goroutine on ctx.Done, accept loop with graceful-shutdown error handling")
}

func parseAddr(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "-addr=") {
			return a[len("-addr="):]
		}
	}
	return ":0"
}
```

- [ ] **Step 3:** Create `lessons/21-networking/exercises/cmd/echo-server/main_test.go` (SKELETON):

```go
package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestServerEcho is a SKELETON. Spawn run() in a goroutine on ":0";
// parse the actual port from the stdout log line; dial it; send
// a line; assert uppercased response; cancel ctx to shut down.
func TestServerEcho(t *testing.T) {
	// TODO:
	//   ctx, cancel := context.WithCancel(context.Background())
	//   defer cancel()
	//   var log bytes.Buffer
	//   errCh := make(chan error, 1)
	//   go func() { errCh <- run(ctx, ":0", &log) }()
	//   wait for "listening on ..." line in log
	//   parse the addr (after "listening on ")
	//   net.Dial("tcp", addr)
	//   send "hello\n"; read response; assert "HELLO\n"
	//   cancel()
	//   if err := <-errCh; err != nil → t.Errorf(...)
	_ = bytes.NewReader
	_ = context.Background
	_ = strings.Contains
}
```

- [ ] **Step 4:** Create `lessons/21-networking/solutions/cmd/echo-server/main.go`:

```go
// Package main is the lesson 21 TCP echo server reference implementation.
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/ristkari-dev/go-training/lessons/21-networking/solutions/warmup/lineecho"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := parseAddr(os.Args[1:])

	if err := run(ctx, addr, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, addr string, stdout io.Writer) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	defer l.Close()

	fmt.Fprintf(stdout, "listening on %s\n", l.Addr())

	// Closer goroutine: when ctx is cancelled, close the listener.
	// This unblocks the Accept loop, which then exits.
	go func() {
		<-ctx.Done()
		_ = l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				// Graceful shutdown — listener closed on purpose.
				return nil
			}
			return fmt.Errorf("accept: %w", err)
		}
		go func() {
			defer conn.Close()
			_ = lineecho.LineEcho(conn)
		}()
	}
}

func parseAddr(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "-addr=") {
			return a[len("-addr="):]
		}
	}
	return ":0"
}
```

- [ ] **Step 5:** Create `lessons/21-networking/solutions/cmd/echo-server/main_test.go`:

```go
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestServerEcho spawns the server on an ephemeral port, parses the
// chosen port from the log line, dials it, sends a line, asserts the
// uppercased response, then cancels ctx to shut down.
func TestServerEcho(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture server's "listening on" log line. Use a thread-safe writer
	// because the server writes to it concurrently with our read.
	var (
		mu  sync.Mutex
		buf bytes.Buffer
	)
	w := &syncWriter{mu: &mu, w: &buf}

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, ":0", w)
	}()

	// Wait for the "listening on" line, parse the addr.
	addr := waitForListenAddr(t, &mu, &buf, 2*time.Second)

	// Dial the server, send a line, read response.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	fmt.Fprintln(conn, "hello")
	r := bufio.NewScanner(conn)
	if !r.Scan() {
		t.Fatalf("no response: %v", r.Err())
	}
	if got := r.Text(); got != "HELLO" {
		t.Errorf("response = %q, want %q", got, "HELLO")
	}

	// Cancel ctx → server shuts down.
	cancel()

	// Wait for run to return — should be nil for graceful shutdown.
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return within 2s after cancel")
	}
}

// syncWriter is a thread-safe io.Writer over a bytes.Buffer.
type syncWriter struct {
	mu *sync.Mutex
	w  *bytes.Buffer
}

func (sw *syncWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

// waitForListenAddr polls the buffer until it sees "listening on <addr>"
// and returns the addr. Fails the test if timeout elapses first.
func waitForListenAddr(t *testing.T, mu *sync.Mutex, buf *bytes.Buffer, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mu.Lock()
		content := buf.String()
		mu.Unlock()

		if idx := strings.Index(content, "listening on "); idx >= 0 {
			rest := content[idx+len("listening on "):]
			if nlIdx := strings.IndexByte(rest, '\n'); nlIdx >= 0 {
				return strings.TrimSpace(rest[:nlIdx])
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server didn't log 'listening on ...' within %v", timeout)
	return ""
}
```

> Note on the `syncWriter` wrapper: `bytes.Buffer` is NOT safe for concurrent use. The server's `run` writes to the buffer; the test's polling reads from it. Without the mutex, race detector flags it. The wrapper is ~10 lines and removes the race.

> Note on `waitForListenAddr`: the server logs `listening on <addr>` right after `net.Listen` succeeds. We poll for that line then parse the addr. The 2-second timeout protects against the server failing to start.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/21-networking/
go test ./lessons/21-networking/exercises/cmd/echo-server/... 2>&1 | tail -5
go test ./lessons/21-networking/solutions/cmd/echo-server/... -v 2>&1 | tail -10
go test -race ./lessons/21-networking/solutions/cmd/echo-server/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

# Smoke test the server (manual; user runs it)
# Terminal 1: go run ./lessons/21-networking/solutions/cmd/echo-server -addr=:9876
# Terminal 2: echo "hello world" | nc localhost 9876   # should print "HELLO WORLD"

git add lessons/21-networking/exercises/cmd/echo-server/ lessons/21-networking/solutions/cmd/echo-server/
git commit -m "feat(lesson-21): cmd/echo-server (TCP listen + accept + per-connection goroutine)"
```

Expected: exercises vacuous-pass; solutions TestServerEcho PASS; -race clean.

---

## Task 4: cmd/echo-client

Dials addr, reads stdin lines, sends to server, prints responses.

**Files (4 total):**

- [ ] **Step 1:** `mkdir -p lessons/21-networking/{exercises,solutions}/cmd/echo-client`

- [ ] **Step 2-5:** Create `lessons/21-networking/{exercises,solutions}/cmd/echo-client/{main.go, main_test.go}`. Same structure as echo-server:

**main.go shape**:
```go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()
    addr := parseAddr(os.Args[1:])
    if err := run(ctx, addr, os.Stdin, os.Stdout); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
    conn, err := net.Dial("tcp", addr)
    if err != nil { return fmt.Errorf("dial %s: %w", addr, err) }
    defer conn.Close()

    // Reader goroutine: read responses from conn, write to stdout.
    go func() {
        r := bufio.NewScanner(conn)
        for r.Scan() {
            fmt.Fprintln(stdout, r.Text())
        }
    }()

    // Main loop: read lines from stdin, send to conn.
    r := bufio.NewScanner(stdin)
    w := bufio.NewWriter(conn)
    for r.Scan() {
        select {
        case <-ctx.Done():
            return nil
        default:
        }
        if _, err := fmt.Fprintln(w, r.Text()); err != nil { return err }
        if err := w.Flush(); err != nil { return err }
    }
    return r.Err()
}
```

**main_test.go shape**: spawn a tiny in-process server (using the same accept-loop + LineEcho pattern, ~15 lines inline), then call `run(ctx, addr, strings.NewReader("hi\n"), &stdout)`, assert stdout contains "HI".

Reuse `parseAddr` from echo-server (cmd binaries are separate packages, so it'll be duplicated — small enough to be acceptable).

Full client test code is similar to TestServerEcho but spawns its own tiny server. Key bits:

```go
func TestClientEcho(t *testing.T) {
	// Spawn a tiny in-process server on :0.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { t.Fatal(err) }
	defer l.Close()

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for {
			conn, err := l.Accept()
			if err != nil { return }
			go func() {
				defer conn.Close()
				_ = lineecho.LineEcho(conn)
			}()
		}
	}()

	// Run the client.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	stdin := strings.NewReader("hi\nworld\n")
	if err := run(ctx, l.Addr().String(), stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
	}

	// Give the reader goroutine a moment to write responses.
	time.Sleep(50 * time.Millisecond)

	out := stdout.String()
	if !strings.Contains(out, "HI") || !strings.Contains(out, "WORLD") {
		t.Errorf("stdout missing expected responses: %q", out)
	}
}
```

> The client test depends on lineecho — import it from `solutions/warmup/lineecho` (or `exercises/...` for the exercises tree).

> The `time.Sleep(50 * time.Millisecond)` for the reader goroutine is a small flake risk; alternative is to use a done channel from the reader to signal completion. For the lesson's scope, the sleep is acceptable; flag it as a known soft-spot.

- [ ] **Step 6:** Verify and commit:

```bash
gofmt -l lessons/21-networking/
go test ./lessons/21-networking/exercises/cmd/echo-client/... 2>&1 | tail -5
go test ./lessons/21-networking/solutions/cmd/echo-client/... -v 2>&1 | tail -10
go test -race ./lessons/21-networking/solutions/cmd/echo-client/... 2>&1 | tail -5
make test
make test-race
golangci-lint run ./...
go vet ./...

git add lessons/21-networking/exercises/cmd/echo-client/ lessons/21-networking/solutions/cmd/echo-client/
git commit -m "feat(lesson-21): cmd/echo-client (net.Dial + stdin/stdout pipe through conn)"
```

---

## Task 5: Slide deck — 4 concepts

Heavy-explanatory pattern per concept. ~28 slides.

**File:** `lessons/21-networking/slides/slides.md`

Concept order:

1. **TCP server** — net.Listen + accept loop + per-connection goroutine. Common-mistake: forgetting `defer conn.Close()` → FD leak.
2. **TCP client** — net.Dial + treating conn as io.ReadWriter. Common-mistake: no deadlines on long-lived conns.
3. **Signals + graceful shutdown** — signal.NotifyContext + listener.Close → Accept errors → loop exits.
4. **Syscalls + file descriptors** — conns/files/pipes are all FDs; syscall package; ulimit -n.

- [ ] **Step 1-7:** Author the deck.
- [ ] **Step 8:** Verify slides build + commit:

```bash
make slides-build
grep -q "21-networking" dist/index.html && echo "✓ in index"
rm -rf dist

git add lessons/21-networking/slides/
git commit -m "feat(lesson-21): slides — networking & syscalls (4 concepts)"
```

---

## Task 6: README + e2e verify + final review + PR

**README**: ~220 lines. Mirrors 4 concepts. Features `make test-race`. Per-concept Common-mistake paragraphs. "How to run" includes manual smoke test instructions (terminal 1 server, terminal 2 nc/client).

After commit, run standard sweep + final review subagent + push branch + create PR.

```bash
make test
make test-race
go test -v ./lessons/21-networking/solutions/...
go test -race ./lessons/21-networking/...
go vet ./...
golangci-lint run ./...
gofmt -l lessons/21-networking/
make slides-build && grep -q "21-networking" dist/index.html && rm -rf dist
```

Dispatch feature-dev:code-reviewer over `git diff main..HEAD`. Apply any fixes. Then push + `gh pr create`.

---

## Critical file paths

To be created:

- `lessons/21-networking/` (directory)
- `lessons/21-networking/README.md`
- `lessons/21-networking/slides/{index.html, slides.md, assets/.gitkeep}`
- `lessons/21-networking/exercises/warmup/lineecho/{lineecho.go, lineecho_test.go}`
- `lessons/21-networking/exercises/cmd/echo-server/{main.go, main_test.go}`
- `lessons/21-networking/exercises/cmd/echo-client/{main.go, main_test.go}`
- `lessons/21-networking/solutions/...` (mirrored)

No modifications to `tools/build-index/main.go` (slug `networking` matches), `Makefile`, or `.golangci.yml`.
