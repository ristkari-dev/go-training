package greeter

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/warmup/greeter/greeterpb"
)

func TestGreet(t *testing.T) {
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	pb.RegisterGreeterServer(srv, &Server{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

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
	defer conn.Close()

	reply, err := pb.NewGreeterClient(conn).Greet(context.Background(), &pb.GreetRequest{Name: "World"})
	if err != nil {
		t.Fatalf("Greet: %v", err)
	}
	if reply.GetMessage() != "Hello, World!" {
		t.Errorf("Greet = %q, want %q", reply.GetMessage(), "Hello, World!")
	}
}
