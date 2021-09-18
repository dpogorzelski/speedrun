package main

import (
	"net"
	"os"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/text"
	"github.com/speedrunsh/speedrun/pkg/portal"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"google.golang.org/grpc"
)

const addr = "0.0.0.0:1337"

func main() {
	h := loghandler.New(os.Stdout)
	log.SetHandler(h)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	portalpb.RegisterPortalServer(s, &portal.Server{})
	log.Infof("Started portal on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
