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
//
//	echo-server -addr=:8080    # fixed port
//	echo-server -addr=:0       # ephemeral port (default; logged after bind)
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
//  1. l, err := net.Listen("tcp", addr)
//  2. if err != nil → return err
//  3. defer l.Close()
//  4. fmt.Fprintf(w, "listening on %s\n", l.Addr())
//  5. Closer goroutine: <-ctx.Done(); l.Close()
//  6. Accept loop:
//     for {
//     conn, err := l.Accept()
//     if err != nil {
//     if ctx.Err() != nil { return nil }  // graceful shutdown
//     return err
//     }
//     go func() {
//     defer conn.Close()
//     _ = lineecho.LineEcho(conn)
//     }()
//     }
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
