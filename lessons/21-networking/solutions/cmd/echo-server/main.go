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
