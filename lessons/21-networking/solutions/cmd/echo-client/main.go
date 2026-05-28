// Package main is the lesson 21 TCP echo client reference implementation.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := parseAddr(os.Args[1:])
	if addr == "" {
		fmt.Fprintln(os.Stderr, "error: -addr=host:port is required")
		os.Exit(2)
	}

	if err := run(ctx, addr, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	// Reader goroutine: print server responses as they arrive.
	// Closes readerDone on exit so the main path can wait for drain.
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		r := bufio.NewScanner(conn)
		for r.Scan() {
			fmt.Fprintln(stdout, r.Text())
		}
	}()

	// Main loop: read stdin lines, send to server (with ctx check).
	r := bufio.NewScanner(stdin)
	w := bufio.NewWriter(conn)
	for r.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if _, err := fmt.Fprintln(w, r.Text()); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("flush: %w", err)
		}
	}
	if err := r.Err(); err != nil {
		return err
	}

	// Half-close the send side so the server sees EOF, finishes
	// responding, and closes its side. Then drain the reader.
	if tc, ok := conn.(*net.TCPConn); ok {
		_ = tc.CloseWrite()
	}
	<-readerDone
	return nil
}

func parseAddr(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "-addr=") {
			return a[len("-addr="):]
		}
	}
	return ""
}
