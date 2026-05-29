package main

import (
	"net/http/httptest"
	"testing"
)

// TestIngestThenStats is a SKELETON. Build the router, POST /ingest with
// some lines, GET /stats, assert the counts.
func TestIngestThenStats(t *testing.T) {
	// TODO:
	//   srv := httptest.NewServer(newRouter(logstats.NewStore(), discardLogger))
	//   defer srv.Close()
	//   POST {"lines":[...]} to srv.URL+"/ingest"; assert accepted/parsed/failed
	//   GET srv.URL+"/stats"; assert counts + total
	_ = httptest.NewServer
}
