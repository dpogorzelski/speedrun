package portal

import (
	"context"
	"os/exec"

	"github.com/apex/log"

	"github.com/speedrunsh/speedrun/proto/portal"
)

func (s *Server) RunCommand(ctx context.Context, in *portal.Command) (*portal.Response, error) {
	fields := log.Fields{
		"context": "command",
	}
	log := log.WithFields(fields)

	log.Debugf("Received command: %s %s", in.GetName(), in.GetArgs())
	cmd := exec.Command(in.GetName(), in.GetArgs()...)
	stdout, err := cmd.Output()

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &portal.Response{Content: string(stdout)}, nil
}
