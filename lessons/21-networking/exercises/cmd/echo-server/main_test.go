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
