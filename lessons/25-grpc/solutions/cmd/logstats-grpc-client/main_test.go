package main

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/internal/grpcsrv"
	"github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/internal/logstats"
	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/proto/logstatspb"
)

func TestRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	pb.RegisterLogStatsServer(srv, grpcsrv.New(logstats.NewStore()))
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out bytes.Buffer
	in := strings.NewReader("2026-01-02T15:04:05 INFO ok\n2026-01-02T15:04:06 ERROR boom\nbad\n")
	if err := run(ctx, lis.Addr().String(), in, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "ingested: accepted=3 parsed=2 failed=1") {
		t.Errorf("missing ingest summary: %q", s)
	}
	if !strings.Contains(s, "total=2") {
		t.Errorf("missing stats: %q", s)
	}
}
