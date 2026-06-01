// Package grpcsrv implements the LogStats gRPC service over the shared
// logstats.Store — the same accumulator the HTTP face uses.
package grpcsrv

import (
	"context"
	"io"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/proto/logstatspb"
)

// Server implements the generated LogStatsServer.
type Server struct {
	pb.UnimplementedLogStatsServer // forward-compat: unknown RPCs default to Unimplemented
	store                          *logstats.Store
}

func New(store *logstats.Store) *Server { return &Server{store: store} }

// GetStats is the unary RPC: snapshot the store into a StatsReply.
// IMPLEMENT THIS.
//
// Hint: counts, total := s.store.Snapshot(); convert map[string]int →
// map[string]int32; return &pb.StatsReply{Counts: out, Total: int32(total)}.
func (s *Server) GetStats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsReply, error) {
	_ = s.store
	panic("TODO: Snapshot the store, convert counts to int32, return StatsReply")
}

// Ingest is the client-streaming RPC: Recv lines until io.EOF, parse +
// accumulate (lenient — count failures), then SendAndClose a summary.
// IMPLEMENT THIS.
//
// Hint:
//
//	delta := map[string]int{}; var accepted, parsed, failed int32
//	for {
//	    line, err := stream.Recv()
//	    if err == io.EOF { s.store.Merge(delta); return stream.SendAndClose(&pb.IngestSummary{...}) }
//	    if err != nil { return err }
//	    accepted++
//	    if e, perr := logparse.ParseLine(line.GetLine()); perr == nil { delta[e.Level]++; parsed++ } else { failed++ }
//	}
func (s *Server) Ingest(stream pb.LogStats_IngestServer) error {
	_ = io.EOF
	_ = logparse.ParseLine
	panic("TODO: Recv loop until io.EOF; lenient parse+count; Merge; SendAndClose(summary)")
}

// LoggingUnaryInterceptor logs each unary call. IMPLEMENT THIS.
//
// Hint: return a grpc.UnaryServerInterceptor that calls handler(ctx, req),
// logs method (info.FullMethod) + err, returns (resp, err).
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	_ = logger
	panic("TODO: return an interceptor that calls handler then logs info.FullMethod + err")
}

// LoggingStreamInterceptor logs each streaming call. IMPLEMENT THIS.
func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	_ = logger
	panic("TODO: return an interceptor that calls handler then logs info.FullMethod + err")
}
