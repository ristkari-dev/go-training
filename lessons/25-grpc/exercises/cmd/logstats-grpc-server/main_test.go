package main

import "testing"

// The gRPC server wiring is exercised end-to-end by the grpc-client
// integration test and by internal/grpcsrv's bufconn tests. This file
// exists so the package has a test target.
func TestServerBuilds(t *testing.T) {}
