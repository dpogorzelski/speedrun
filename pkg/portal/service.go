package portal

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/speedrunsh/speedrun/proto/portal"
)

func (s *Server) ServiceRestart(ctx context.Context, service *portal.Service) (*portal.Response, error) {
	fields := log.Fields{
		"context": "service",
		"command": "restart",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", service.GetName())
	_, err = conn.RestartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	res := <-responseChan
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStop(ctx context.Context, service *portal.Service) (*portal.Response, error) {
	fields := log.Fields{
		"context": "service",
		"command": "stop",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", service.GetName())
	_, err = conn.StopUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	res := <-responseChan
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStart(ctx context.Context, service *portal.Service) (*portal.Response, error) {
	fields := log.Fields{
		"context": "service",
		"command": "start",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", service.GetName())
	_, err = conn.StartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	res := <-responseChan
	return &portal.Response{Content: res}, nil

}

func (s *Server) ServiceStatus(ctx context.Context, service *portal.Service) (*portal.ServiceStatusResponse, error) {
	fields := log.Fields{
		"context": "service",
		"command": "status",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)
	log.Debug("Received service status request")

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	serviceName := fmt.Sprintf("%s.service", service.GetName())
	res, err := conn.ListUnitsByNamesContext(ctx, []string{serviceName})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debugf("Fetched service list by name: %v", res)
	if res[0].LoadState == "not-found" {
		log.Error("service not found")
		return nil, fmt.Errorf("service not found")
	}

	return &portal.ServiceStatusResponse{
		ActiveState: res[0].ActiveState,
		LoadState:   res[0].LoadState,
		SubState:    res[0].SubState,
	}, nil

}
