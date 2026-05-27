package main

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ristkari-dev/go-training/lessons/21-networking/solutions/warmup/lineecho"
)

// TestClientEcho spawns a tiny in-process server on :0, runs the
// client against it, asserts stdout contains the uppercased responses.
// Because run() now drains the reader goroutine before returning
// (via CloseWrite() + <-readerDone), no sleep is needed.
func TestClientEcho(t *testing.T) {
	// Tiny in-process server using lineecho.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer l.Close()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				_ = lineecho.LineEcho(conn)
			}()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var stdout bytes.Buffer
	stdin := strings.NewReader("hi\nworld\n")
	if err := run(ctx, l.Addr().String(), stdin, &stdout); err != nil {
		t.Fatalf("run: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "HI") || !strings.Contains(out, "WORLD") {
		t.Errorf("stdout missing expected responses: %q", out)
	}
}
