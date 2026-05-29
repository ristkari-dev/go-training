// Package main is the lesson 25 LogStats gRPC client. It streams stdin
// lines to the server via the Ingest client-streaming RPC, prints the
// summary, then calls the unary GetStats.
//
// Usage: logstats-grpc-client -addr=127.0.0.1:9090 < app.log
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/proto/logstatspb"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := flag.String("addr", "127.0.0.1:9090", "gRPC server address")
	flag.Parse()

	if err := run(ctx, *addr, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewLogStatsClient(conn)

	// Stream stdin lines via the client-streaming Ingest RPC.
	stream, err := client.Ingest(ctx)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(stdin)
	for sc.Scan() {
		if err := stream.Send(&pb.LogLine{Line: sc.Text()}); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	summary, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "ingested: accepted=%d parsed=%d failed=%d\n",
		summary.GetAccepted(), summary.GetParsed(), summary.GetFailed())

	// Then the unary GetStats.
	reply, err := client.GetStats(ctx, &pb.StatsRequest{})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "stats: total=%d counts=%v\n", reply.GetTotal(), reply.GetCounts())
	return nil
}
