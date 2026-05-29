// Package main is the lesson 24 logstats client: it reads log lines
// from stdin (or -file) and ships them to a logstats server's /ingest
// using the resilient shipper (timeouts, selective retry, breaker).
//
// Usage:
//
//	logstats-client -addr=http://localhost:8080/ingest < app.log
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/ristkari-dev/go-training/lessons/24-http-client/exercises/internal/shipper"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := flag.String("addr", "http://localhost:8080/ingest", "ingest endpoint URL")
	flag.Parse()

	if err := run(ctx, *addr, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run reads all lines from stdin, ships them in one batch, and prints
// the server's summary. (A production client would batch by size and
// stream; one batch keeps the lesson focused on resilience.)
func run(ctx context.Context, addr string, stdin io.Reader, stdout io.Writer) error {
	var lines []string
	sc := bufio.NewScanner(stdin)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return err
	}

	res, err := shipper.New(addr).Ship(ctx, lines)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "shipped: accepted=%d parsed=%d failed=%d\n", res.Accepted, res.Parsed, res.Failed)
	return nil
}
