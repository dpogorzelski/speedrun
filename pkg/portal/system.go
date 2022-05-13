//go:build linux && amd64

package portal

import (
	"context"
	"syscall"

	"github.com/apex/log"
	"github.com/dpogorzelski/speedrun/proto/portal"
)

func (s *Server) SystemReboot(ctx context.Context, system *portal.SystemRebootRequest) (*portal.SystemRebootResponse, error) {
	fields := log.Fields{
		"context": "system",
		"command": "reboot",
	}
	log := log.WithFields(fields)
	log.Debug("Received system reboot request")

	syscall.Sync()
	go syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)

	return &portal.SystemRebootResponse{State: portal.State_CHANGED, Message: "Rebooting"}, nil
}

func (s *Server) SystemShutdown(ctx context.Context, system *portal.SystemShutdownRequest) (*portal.SystemShutdownResponse, error) {
	fields := log.Fields{
		"context": "system",
		"command": "shutdown",
	}
	log := log.WithFields(fields)
	log.Debug("Received system shutdown request")

	syscall.Sync()
	go syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)

	return &portal.SystemShutdownResponse{State: portal.State_CHANGED, Message: "Shutting down"}, nil
}
