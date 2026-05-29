package greeter

import (
	"testing"

	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/warmup/greeter/greeterpb"
)

// TestGreet is a SKELETON. Stand up the Greeter over bufconn, call
// Greet("World"), assert "Hello, World!".
func TestGreet(t *testing.T) {
	// TODO: bufconn.Listen → grpc.NewServer → RegisterGreeterServer(&Server{})
	//       → grpc.NewClient(passthrough, WithContextDialer, insecure)
	//       → NewGreeterClient.Greet(ctx, &pb.GreetRequest{Name:"World"})
	//       → assert reply.GetMessage() == "Hello, World!"
	_ = pb.NewGreeterClient
}
