package grpcsrv

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/ristkari-dev/go-training/lessons/26-config/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/26-config/solutions/proto/logstatspb"
)

func dialer(t *testing.T) (pb.LogStatsClient, func()) {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(LoggingUnaryInterceptor(logger)),
		grpc.StreamInterceptor(LoggingStreamInterceptor(logger)),
	)
	pb.RegisterLogStatsServer(srv, New(logstats.NewStore()))
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return pb.NewLogStatsClient(conn), func() { conn.Close(); srv.Stop() }
}

func TestGetStatsEmpty(t *testing.T) {
	client, cleanup := dialer(t)
	defer cleanup()
	reply, err := client.GetStats(context.Background(), &pb.StatsRequest{})
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if reply.GetTotal() != 0 {
		t.Errorf("total = %d, want 0", reply.GetTotal())
	}
}

func TestIngestThenGetStats(t *testing.T) {
	client, cleanup := dialer(t)
	defer cleanup()

	stream, err := client.Ingest(context.Background())
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	for _, l := range []string{
		"2026-01-02T15:04:05 INFO ok",
		"2026-01-02T15:04:06 INFO again",
		"2026-01-02T15:04:07 ERROR boom",
		"garbage",
	} {
		if err := stream.Send(&pb.LogLine{Line: l}); err != nil {
			t.Fatalf("Send: %v", err)
		}
	}
	summary, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("CloseAndRecv: %v", err)
	}
	if summary.GetAccepted() != 4 || summary.GetParsed() != 3 || summary.GetFailed() != 1 {
		t.Errorf("summary = %+v, want accepted=4 parsed=3 failed=1", summary)
	}

	reply, err := client.GetStats(context.Background(), &pb.StatsRequest{})
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if reply.GetTotal() != 3 || reply.GetCounts()["INFO"] != 2 || reply.GetCounts()["ERROR"] != 1 {
		t.Errorf("stats = %+v", reply)
	}
}
