package main

import (
	"log"
	"net"

	"github.com/speedrunsh/portal-api/go/service"
	"github.com/speedrunsh/speedrun/portal"

	"google.golang.org/grpc"
)

const addr = "0.0.0.0:1337"

func main() {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	service.RegisterPortalServer(s, &portal.Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
