package grpcsrv

import (
	"testing"

	pb "github.com/ristkari-dev/go-training/lessons/29-capstone-final/exercises/proto/logstatspb"
)

// TestService is a SKELETON. Stand up the LogStats service over bufconn
// (with the interceptors), then: GetStats on empty → total 0; Ingest a
// few lines → summary; GetStats → counts. See the solution for the
// bufconn dialer pattern.
func TestService(t *testing.T) {
	// TODO: bufconn + grpc.NewServer(UnaryInterceptor, StreamInterceptor)
	//       + RegisterLogStatsServer(New(logstats.NewStore())) + client calls.
	_ = pb.NewLogStatsClient
}
