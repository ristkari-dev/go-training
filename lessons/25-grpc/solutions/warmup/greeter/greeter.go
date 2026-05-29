// Package greeter is the lesson 25 warm-up reference implementation.
package greeter

import (
	"context"

	pb "github.com/ristkari-dev/go-training/lessons/25-grpc/solutions/warmup/greeter/greeterpb"
)

type Server struct {
	pb.UnimplementedGreeterServer
}

func (s *Server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetReply, error) {
	return &pb.GreetReply{Message: "Hello, " + req.GetName() + "!"}, nil
}
