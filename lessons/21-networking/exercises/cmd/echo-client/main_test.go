package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestClientEcho is a SKELETON. Spin up a tiny in-process server,
// run() the client against it, assert stdout contains the uppercased
// responses. Because run() now waits for the reader goroutine to
// drain (via the readerDone + CloseWrite pattern), the assertion
// can read stdout right after run() returns — no sleep needed.
func TestClientEcho(t *testing.T) {
	// TODO:
	//   l, _ := net.Listen("tcp", "127.0.0.1:0"); defer l.Close()
	//   go func() {
	//       for {
	//           conn, err := l.Accept(); if err != nil { return }
	//           go func() { defer conn.Close(); _ = lineecho.LineEcho(conn) }()
	//       }
	//   }()
	//   ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//   defer cancel()
	//   var stdout bytes.Buffer
	//   stdin := strings.NewReader("hi\nworld\n")
	//   if err := run(ctx, l.Addr().String(), stdin, &stdout); err != nil { t.Fatal(err) }
	//   assert stdout contains "HI" and "WORLD"
	_ = bytes.NewReader
	_ = context.Background
	_ = strings.Contains
}
