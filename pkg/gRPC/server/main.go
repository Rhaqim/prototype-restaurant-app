package main

import (
	"context"
	"flag"
	"log"
	"net"

	pb "github.com/Rhaqim/thedutchapp/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/alts"
)

var (
	addr = flag.String("addr", ":50051", "listen address")
)

type server struct {
	pb.UnimplementedDutchServiceServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening on %s", *addr)
	altsTC := alts.NewServerCreds(alts.DefaultServerOptions())
	s := grpc.NewServer(grpc.Creds(altsTC))
	pb.RegisterDutchServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Printf("Server started")
}

func (s *server) SignIn(ctx context.Context, in *pb.SignInRequest) (*pb.SignInResponse, error) {
	log.Printf("ComputeSignin: %v", in.GetUsername())
	return &pb.SignInResponse{Message: "Hello " + in.GetUsername()}, nil
}

func (s *server) SignUp(ctx context.Context, in *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	log.Printf("ComputeSignup: %v", in.GetUsername())
	return &pb.SignUpResponse{Message: "Hello " + in.GetUsername()}, nil
}
