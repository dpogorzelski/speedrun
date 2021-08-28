package portal

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	qnet "github.com/speedrunsh/grpc-quic"

	"github.com/speedrunsh/speedrun/key"
	"github.com/speedrunsh/speedrun/ssh"
	"google.golang.org/grpc"
)

func SSHTransport(address string, key *key.Key) (*grpc.ClientConn, error) {
	sshclient, err := ssh.Connect(address, key)
	if err != nil {
		return nil, err
	}

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return sshclient.Dial("tcp", "127.0.0.1:1337")
	}

	return grpc.Dial("127.0.0.1:1337", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
}

func SSHTransportInsecure(address string, key *key.Key) (*grpc.ClientConn, error) {
	sshclient, err := ssh.ConnectInsecure(address, key)
	if err != nil {
		return nil, err
	}

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return sshclient.Dial("tcp", "127.0.0.1:1337")
	}

	return grpc.Dial("127.0.0.1:1337", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
}

func HTTP2Transport(address string) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("%s:%d", address, 1337)
	return grpc.Dial(target, grpc.WithInsecure())
}

func QUICTransport(address string) (*grpc.ClientConn, error) {
	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		CurvePreferences:   []tls.CurveID{tls.X25519},
		CipherSuites:       []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		InsecureSkipVerify: true,
		NextProtos:         []string{"speedrun"},
	}

	creds := qnet.NewCredentials(tlsConf)

	dialer := qnet.NewQuicDialer(tlsConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	target := fmt.Sprintf("%s:%d", address, 1337)
	return grpc.Dial(target, grpcOpts...)
}
