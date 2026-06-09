// Package grpcsrv implements the LogStats gRPC service over the shared
// logstats.Store — the same accumulator the HTTP face uses. Reference impl.
package grpcsrv

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logparse"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/proto/logstatspb"
)

type Server struct {
	pb.UnimplementedLogStatsServer
	store    *logstats.Store
	ingested metric.Int64Counter
}

func New(store *logstats.Store) *Server {
	// Same ingested_lines_total metric as the HTTP face — both transports
	// record into one counter via the global MeterProvider.
	ingested, _ := otel.Meter("logstatsd").Int64Counter("ingested_lines_total")
	return &Server{store: store, ingested: ingested}
}

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
			s.ingested.Add(stream.Context(), int64(parsed+failed))
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

// LoggingUnaryInterceptor traces + logs each unary call (gRPC's middleware).
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("logstatsd/grpc")
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()
		resp, err := handler(ctx, req)
		logger.Info("unary", "method", info.FullMethod, "err", err)
		return resp, err
	}
}

// LoggingStreamInterceptor traces + logs each streaming call. It wraps the
// stream so the span's context flows into the handler.
func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	tracer := otel.Tracer("logstatsd/grpc")
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, span := tracer.Start(ss.Context(), info.FullMethod)
		defer span.End()
		err := handler(srv, &tracingServerStream{ServerStream: ss, ctx: ctx})
		logger.Info("stream", "method", info.FullMethod, "err", err)
		return err
	}
}

// tracingServerStream overrides Context so the handler sees the span's ctx.
type tracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (t *tracingServerStream) Context() context.Context { return t.ctx }
