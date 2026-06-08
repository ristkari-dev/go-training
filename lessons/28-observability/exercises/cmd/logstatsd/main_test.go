package main

import "testing"

// TestServe is a SKELETON. Stand up two 127.0.0.1:0 listeners + an
// http.Server (httpsrv.Router) + a grpc.Server (grpcsrv registered),
// run serve in a goroutine, hit /healthz over HTTP and GetStats over
// gRPC, cancel the ctx, and assert serve returns nil + both stopped.
func TestServe(t *testing.T) {
	// TODO — see the solution.
}
