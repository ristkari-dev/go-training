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
