package portal

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/speedrunsh/speedrun/proto/portal"
)

func (s *Server) ServiceRestart(ctx context.Context, in *portal.Service) (*portal.Response, error) {
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
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStop(ctx context.Context, in *portal.Service) (*portal.Response, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", in.GetName())
	_, err = conn.StopUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return nil, err
	}

	res := <-responseChan
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStart(ctx context.Context, in *portal.Service) (*portal.Response, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", in.GetName())
	_, err = conn.StartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		return nil, err
	}

	res := <-responseChan
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStatus(ctx context.Context, in *portal.Service) (*portal.Response, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	serviceName := fmt.Sprintf("%s.service", in.GetName())
	res, err := conn.ListUnitsByNamesContext(ctx, []string{serviceName})
	if err != nil {
		return nil, err
	}

	return &portal.Response{Content: res[0].ActiveState}, nil

}
