package httpsrv

import (
	"net/http/httptest"
	"testing"
)

// TestRouter is a SKELETON. Router itself is provided (working), but the
// carried logstats.Store in this exercises tree is still a skeleton
// (Merge/Snapshot panic until you implement it), so exercising the full
// HTTP ingest/stats path here would panic. The real handler tests live
// in the solutions tree. Once you've implemented logstats, flesh this
// out: stand up Router over httptest, POST /ingest, GET /stats, assert.
func TestRouter(t *testing.T) {
	// TODO: srv := httptest.NewServer(Router(logstats.NewStore(), dedup.New(1000), discardLogger))
	//       POST /ingest; GET /stats; assert counts. (dedup.Seen is a
	//       skeleton here, so don't set an Idempotency-Key until it's done.)
	_ = httptest.NewServer
}
