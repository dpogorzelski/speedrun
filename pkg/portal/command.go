package portal

import (
	"context"
	"os/exec"
	"strings"

	"github.com/apex/log"

	"github.com/dpogorzelski/speedrun/proto/portal"
)

func (s *Server) RunCommand(ctx context.Context, in *portal.CommandRequest) (*portal.CommandResponse, error) {
	fields := log.Fields{
		"context": "command",
	}
	log := log.WithFields(fields)

	log.Debugf("Received command: %s %s", in.GetName(), in.GetArgs())
	cmd := exec.Command(in.GetName(), in.GetArgs()...)
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &portal.CommandResponse{Message: strings.TrimSpace(string(stdout))}, nil
}
