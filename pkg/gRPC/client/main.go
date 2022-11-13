package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/alts"

	pb "github.com/Rhaqim/thedutchapp/pkg/gRPC"
)

var (
	addr = flag.String("addr", ":50051", "listen address")
	// data   = flag.String("data", "", "Data file to read from")
	// first  = flag.Int64("first", 0, "First number to add")
	// second = flag.Int64("second", 0, "Second number to add")
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	flag.Parse()
	clientOpts := alts.DefaultClientOptions()
	// clientOpts.TargetServiceAccounts = []string{"default"}
	altsTC := alts.NewClientCreds(clientOpts)
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(altsTC))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDutchServiceClient(conn)
	r, err := c.SignIn(ctx, &pb.SignInRequest{Username: "rhaqim", Password: "123"})
	if err != nil {
		log.Fatalf("could not compute checksum: %v", err)
	}
	log.Printf("Checksum: %x", r.Message)

	read, err := c.SignUp(ctx, &pb.SignUpRequest{Username: "rhaqim", Password: "123"})
	if err != nil {
		log.Fatalf("could not compute checksum: %v", err)
	}
	log.Printf("Checksum: %x", read.Message)
}
