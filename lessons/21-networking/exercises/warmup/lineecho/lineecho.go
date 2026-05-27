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
//  1. r := bufio.NewScanner(conn)
//  2. w := bufio.NewWriter(conn)
//  3. for r.Scan() {
//     upper := strings.ToUpper(r.Text())
//     if _, err := fmt.Fprintln(w, upper); err != nil { return err }
//     if err := w.Flush(); err != nil { return err }    // flush per line
//     }
//  4. return r.Err()    // nil on clean EOF
//
// Flush per line is essential for an interactive protocol — the client
// is waiting for a response after each line, so the response must
// actually be sent (not sit in the buffer).
func LineEcho(conn net.Conn) error {
	_ = bufio.NewScanner
	_ = strings.ToUpper
	panic("TODO: bufio.Scanner read loop + bufio.Writer per-line flush; return scanner.Err()")
}
