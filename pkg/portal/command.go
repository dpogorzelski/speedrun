package portal

import (
	"context"

	"github.com/apex/log"

	"github.com/speedrunsh/speedrun/proto/portal"
)

func (s *Server) RunCommand(ctx context.Context, in *portal.Command) (*portal.Response, error) {
	log.Infof("Received command:%s", in.GetName())
	return &portal.Response{Content: "ran " + in.GetName()}, nil
}
