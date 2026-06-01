// Package greeter is the lesson 25 warm-up: a minimal unary gRPC
// service. The .proto + generated code (greeterpb) are provided; you
// implement the one server method.
package greeter

import (
	"context"

	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/exercises/warmup/greeter/greeterpb"
)

// Server implements the generated GreeterServer.
type Server struct {
	pb.UnimplementedGreeterServer
}

// Greet returns "Hello, <name>!". IMPLEMENT THIS.
//
// Hint: return &pb.GreetReply{Message: "Hello, " + req.GetName() + "!"}, nil
func (s *Server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetReply, error) {
	_ = req
	panic("TODO: return a GreetReply with Message \"Hello, <name>!\"")
}
