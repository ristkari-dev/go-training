package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/httpsrv"
	"github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/29-capstone-final/solutions/proto/logstatspb"
)

func TestServe(t *testing.T) {
	httpLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	store := logstats.NewStore()
	httpSrv := &http.Server{Handler: httpsrv.Router(store, logger)}
	grpcSrv := grpc.NewServer()
	pb.RegisterLogStatsServer(grpcSrv, grpcsrv.New(store))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- serve(ctx, httpSrv, httpLis, grpcSrv, grpcLis) }()

	// HTTP face works.
	hurl := fmt.Sprintf("http://%s/healthz", httpLis.Addr().String())
	ready := false
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); {
		if resp, err := http.Get(hurl); err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("HTTP never became ready")
	}

	// gRPC face works.
	conn, err := grpc.NewClient(grpcLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()
	if _, err := pb.NewLogStatsClient(conn).GetStats(context.Background(), &pb.StatsRequest{}); err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	// Cancel → both drain → serve returns nil.
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("serve returned %v, want nil", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("serve did not return after cancel")
	}
	if _, err := http.Get(hurl); err == nil {
		t.Error("HTTP still serving after drain")
	}
}
