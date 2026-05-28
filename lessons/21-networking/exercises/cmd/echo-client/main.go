// Package main is the lesson 21 TCP echo client.
//
// Dials -addr (no default — must be supplied). Reads lines from stdin,
// sends each + "\n" to the server, prints the server's response line
// to stdout. Reader runs in its own goroutine so responses appear as
// they arrive (not after stdin closes).
//
// Half-close pattern: when stdin hits EOF, we call CloseWrite() on
// the TCP conn — that tells the server "no more from me, but keep
// sending me your remaining responses." Then we wait for the reader
// goroutine to drain everything before returning. Closing the whole
// conn too early would discard in-flight responses.
//
// Exit conditions:
//   - stdin EOF (Ctrl-D)   → half-close + drain + clean exit
//   - SIGINT (Ctrl-C)      → ctx cancelled; main loop returns; defer closes conn
//
// Usage:
//
//	echo "hello world" | echo-client -addr=localhost:9876
//	echo-client -addr=localhost:9876    # interactive (type lines)
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

// run is the testable entry point.
//
// Hint:
//  1. conn, err := net.Dial("tcp", addr); defer conn.Close()
//  2. readerDone := make(chan struct{})
//  3. Reader goroutine (closes readerDone on exit):
//     go func() {
//     defer close(readerDone)
//     r := bufio.NewScanner(conn)
//     for r.Scan() { fmt.Fprintln(stdout, r.Text()) }
//     }()
//  4. Main stdin loop:
//     r := bufio.NewScanner(stdin)
//     w := bufio.NewWriter(conn)
//     for r.Scan() {
//     select { case <-ctx.Done(): return nil; default: }
//     if _, err := fmt.Fprintln(w, r.Text()); err != nil { return err }
//     if err := w.Flush(); err != nil { return err }
//     }
//     if err := r.Err(); err != nil { return err }
//  5. Half-close + drain:
//     if tc, ok := conn.(*net.TCPConn); ok { _ = tc.CloseWrite() }
//     <-readerDone
//     return nil
//
// The reader goroutine is essential because the client can't predict
// when responses arrive. Without it, the client would block on Write
// while a response sits unread in the kernel buffer.
//
// CloseWrite() is the TCP half-close: it signals EOF on our send side
// while keeping the receive side open. The server's bufio.Scanner sees
// the EOF, exits its read loop, and closes the conn cleanly. Our
// reader goroutine then sees its own EOF on the receive side, exits,
// and closes readerDone — at which point we return.
func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
	_ = net.Dial
	_ = bufio.NewScanner
	_ = bufio.NewWriter
	panic("TODO: dial; reader goroutine printing responses; stdin loop sending lines (with ctx check); CloseWrite() + drain via readerDone")
}

func parseAddr(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "-addr=") {
			return a[len("-addr="):]
		}
	}
	return ""
}
