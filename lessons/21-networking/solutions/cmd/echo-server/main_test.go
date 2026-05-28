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
