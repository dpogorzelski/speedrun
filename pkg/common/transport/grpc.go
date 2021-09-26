package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/melbahja/goph"
	qnet "github.com/speedrunsh/grpc-quic"

	"github.com/speedrunsh/speedrun/pkg/common/key"
	"github.com/speedrunsh/speedrun/pkg/common/ssh"
	"google.golang.org/grpc"
)

type GRPCTransport struct {
	Conn    *grpc.ClientConn
	address string
	opts    options
}

type options struct {
	insecure bool
	useSSH   bool
	key      *key.Key
}

type TransportOption interface {
	apply(*options)
}

func defaultOptions() options {
	return options{
		insecure: false,
		useSSH:   true,
	}
}

type withInsecure bool

func (w withInsecure) apply(o *options) {
	o.insecure = bool(w)
}

func WithInsecure(enable bool) TransportOption {
	return withInsecure(enable)
}

type withSSH bool

func (w withSSH) apply(o *options) {
	o.insecure = bool(w)
}

func WithSSH(enable bool) TransportOption {
	return withSSH(enable)
}

type withSSHKey key.Key

func (w withSSHKey) apply(o *options) {
	a := key.Key(w)
	o.key = &a
}

func WithSSHKey(key key.Key) TransportOption {
	return withSSHKey(key)
}

func NewGRPCTransport(address string, opts ...TransportOption) (*grpc.ClientConn, error) {
	var err error

	t := &GRPCTransport{
		address: address,
		opts:    defaultOptions(),
	}
	for _, opt := range opts {
		opt.apply(&t.opts)
	}

	if t.opts.useSSH {
		if t.opts.insecure {
			t.Conn, err = sshTransportInsecure(t.address, t.opts.key)
		} else {
			t.Conn, err = sshTransport(t.address, t.opts.key)
		}
		if err != nil {
			return nil, err
		}
	} else {
		if t.opts.insecure {
			t.Conn, err = http2TransportInsecure(address)
			if err != nil {
				return nil, err
			}
		}
	}

	return t.Conn, nil
}

func sshTransportInsecure(address string, key *key.Key) (*grpc.ClientConn, error) {
	var sshclient *goph.Client

	sshclient, err := ssh.ConnectInsecure(address, key)
	if err != nil {
		return nil, err
	}

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return sshclient.Dial("tcp", "127.0.0.1:1337")
	}

	return grpc.Dial("127.0.0.1:1337", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
}

func sshTransport(address string, key *key.Key) (*grpc.ClientConn, error) {
	var sshclient *goph.Client

	sshclient, err := ssh.Connect(address, key)
	if err != nil {
		return nil, err
	}

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return sshclient.Dial("tcp", "127.0.0.1:1337")
	}

	return grpc.Dial("127.0.0.1:1337", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
}

func http2TransportInsecure(address string) (*grpc.ClientConn, error) {
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
