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
		"name":    in.GetName(),
	}
	log := log.WithFields(fields)

	log.Infof("Received command: %s", in.GetName())
	cmd := exec.Command(in.GetName())
	stdout, err := cmd.Output()

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return &portal.Response{Content: string(stdout)}, nil
}
