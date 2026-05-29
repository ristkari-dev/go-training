package main

import (
	"net/http/httptest"
	"testing"
)

// TestRun is a SKELETON. Spin up an httptest server returning an ingest
// response; call run with a strings.Reader of log lines; assert the
// printed summary. (Until shipper.Ship is implemented, run panics — so
// this skeleton only references httptest to keep the import alive.)
func TestRun(t *testing.T) {
	// TODO: httptest server → run(ctx, srv.URL, strings.NewReader(...), &buf)
	//       assert buf contains "accepted=... parsed=... failed=..."
	_ = httptest.NewServer
}
