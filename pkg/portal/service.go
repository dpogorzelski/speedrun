package portal

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/speedrunsh/speedrun/proto/portal"
)

func (s *Server) ServiceRestart(ctx context.Context, service *portal.ServiceRequest) (*portal.ServiceResponse, error) {
	fields := log.Fields{
		"context": "service",
		"command": "restart",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)
	log.Debug("Received service restart request")

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
	log.Debugf("Service restart result: %v", res)
	return &portal.ServiceResponse{Changed: true, Message: strings.Title(res)}, nil

}

func (s *Server) ServiceStop(ctx context.Context, service *portal.ServiceRequest) (*portal.ServiceResponse, error) {
	fields := log.Fields{
		"context": "service",
		"command": "stop",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)
	log.Debug("Received service stop request")

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", service.GetName())
	list, err := conn.ListUnitsByNamesContext(ctx, []string{serviceName})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debugf("Fetched service list by name: %v", list)
	if list[0].ActiveState == "inactive" {
		return &portal.ServiceResponse{Changed: false, Message: "Service already stopped"}, nil
	}

	_, err = conn.StopUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	res := <-responseChan
	log.Debugf("Service stop result: %v", res)
	return &portal.ServiceResponse{Changed: true, Message: strings.Title(res)}, nil

}

func (s *Server) ServiceStart(ctx context.Context, service *portal.ServiceRequest) (*portal.ServiceResponse, error) {
	fields := log.Fields{
		"context": "service",
		"command": "start",
		"name":    service.GetName(),
	}
	log := log.WithFields(fields)
	log.Debug("Received service start request")

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer conn.Close()

	responseChan := make(chan string, 1)
	serviceName := fmt.Sprintf("%s.service", service.GetName())
	list, err := conn.ListUnitsByNamesContext(ctx, []string{serviceName})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	log.Debugf("Fetched service list by name: %v", list)
	if list[0].ActiveState == "active" {
		return &portal.ServiceResponse{Changed: false, Message: "Service already running"}, nil
	}

	_, err = conn.StartUnitContext(ctx, serviceName, "replace", responseChan)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	res := <-responseChan
	log.Debugf("Service start result: %v", res)
	return &portal.ServiceResponse{Changed: true, Message: strings.Title(res)}, nil

}

func (s *Server) ServiceStatus(ctx context.Context, service *portal.ServiceRequest) (*portal.ServiceStatusResponse, error) {
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
		Activestate: res[0].ActiveState,
		Loadstate:   res[0].LoadState,
		Substate:    res[0].SubState,
	}, nil

}
