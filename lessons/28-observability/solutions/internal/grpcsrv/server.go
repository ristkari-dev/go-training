// Package grpcsrv implements the LogStats gRPC service over the shared
// logstats.Store — the same accumulator the HTTP face uses. Reference impl.
package grpcsrv

import (
	"context"
	"io"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/28-observability/solutions/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/28-observability/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/28-observability/solutions/proto/logstatspb"
)

type Server struct {
	pb.UnimplementedLogStatsServer
	store *logstats.Store
}

func New(store *logstats.Store) *Server { return &Server{store: store} }

func (s *Server) GetStats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsReply, error) {
	counts, total := s.store.Snapshot()
	out := make(map[string]int32, len(counts))
	for k, v := range counts {
		out[k] = int32(v)
	}
	return &pb.StatsReply{Counts: out, Total: int32(total)}, nil
}

func (s *Server) Ingest(stream pb.LogStats_IngestServer) error {
	delta := map[string]int{}
	var accepted, parsed, failed int32
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			s.store.Merge(delta)
			return stream.SendAndClose(&pb.IngestSummary{
				Accepted: accepted, Parsed: parsed, Failed: failed,
			})
		}
		if err != nil {
			return err
		}
		accepted++
		if e, perr := logparse.ParseLine(line.GetLine()); perr == nil {
			delta[e.Level]++
			parsed++
		} else {
			failed++
		}
	}
}

// LoggingUnaryInterceptor logs each unary call (gRPC's middleware).
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		logger.Info("unary", "method", info.FullMethod, "err", err)
		return resp, err
	}
}

// LoggingStreamInterceptor logs each streaming call.
func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		logger.Info("stream", "method", info.FullMethod, "err", err)
		return err
	}
}
