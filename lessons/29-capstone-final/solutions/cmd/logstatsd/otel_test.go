package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/proto/logstatspb"
)

// TestGRPCSpan asserts the gRPC interceptor records a span per RPC,
// using core OTel (no otelgrpc contrib). bufconn + in-memory exporter.
func TestGRPCSpan(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)))

	lis := bufconn.Listen(1 << 20)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcsrv.LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(grpcsrv.LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(srv, grpcsrv.New(logstats.NewStore()))
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := pb.NewLogStatsClient(conn).GetStats(context.Background(), &pb.StatsRequest{}); err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	const want = "/logstats.v1.LogStats/GetStats"
	var names []string
	for _, s := range exp.GetSpans() {
		names = append(names, s.Name)
		if s.Name == want {
			return
		}
	}
	t.Errorf("no span named %q; got %v", want, names)
}
