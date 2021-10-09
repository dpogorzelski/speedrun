package main

import (
	"context"
	"crypto/tls"
	"os"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/text"
	itls "github.com/speedrunsh/speedrun/pkg/common/tls"
	"github.com/speedrunsh/speedrun/pkg/portal"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

const addr = "0.0.0.0:1337"

func main() {
	h := loghandler.New(os.Stdout)
	log.SetHandler(h)

	m := drpcmux.New()

	portalpb.DRPCRegisterPortal(m, &portal.Server{})
	log.Infof("Started portal on %s", addr)

	s := drpcserver.New(m)

	tlsConfig, err := itls.GenerateTLSConfig()
	if err != nil {
		log.Fatalf("failed to generate tls config: %v", err)
	}

	lis, err := tls.Listen("tcp", addr, tlsConfig)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	ctx := context.Background()
	if err := s.Serve(ctx, lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
