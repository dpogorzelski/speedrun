package portal

import (
	"context"
	"fmt"
	"log"

	"github.com/speedrunsh/portal-api/go/service"

	"github.com/coreos/go-systemd/v22/dbus"
)

const addr = "0.0.0.0:1337"

type Server struct {
	service.UnimplementedPortalServer
}

func (s *Server) Echo(ctx context.Context, in *service.Empty) (*service.Empty, error) {
	log.Printf("Received ping")
	return &service.Empty{}, nil
}

func (s *Server) ServiceRestart(ctx context.Context, in *service.Service) (*service.Response, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", in.GetName())
	_, err = conn.RestartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return nil, err
	}

	res := <-responseChan
	return &service.Response{Content: res}, nil

}
