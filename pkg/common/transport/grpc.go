package transport

import (
	"crypto/tls"
	"fmt"

	qnet "github.com/speedrunsh/grpc-quic"

	"google.golang.org/grpc"
)

type GRPCTransport struct {
	Conn    *grpc.ClientConn
	address string
	opts    options
}

type options struct {
	insecure bool
}

type TransportOption interface {
	apply(*options)
}

func defaultOptions() options {
	return options{
		insecure: false,
	}
}

type withInsecure bool

func (w withInsecure) apply(o *options) {
	o.insecure = bool(w)
}

func WithInsecure(enable bool) TransportOption {
	return withInsecure(enable)
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

	if t.opts.insecure {
		t.Conn, err = quicTransport(address)
		if err != nil {
			return nil, err
		}
	} else {
		// this should seutp mtls
		t.Conn, err = quicTransport(address)
		if err != nil {
			return nil, err
		}
	}

	if t.Conn == nil {
		return nil, fmt.Errorf("couldn't initialize grpc client")
	}

	return t.Conn, nil
}

func quicTransport(address string) (*grpc.ClientConn, error) {
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
