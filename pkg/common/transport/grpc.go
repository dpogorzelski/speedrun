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
	http2    bool
}

type TransportOption interface {
	apply(*options)
}

func defaultOptions() options {
	return options{
		insecure: false,
		http2:    false,
	}
}

type withInsecure bool

func (w withInsecure) apply(o *options) {
	o.insecure = bool(w)
}

func WithInsecure(enable bool) TransportOption {
	return withInsecure(enable)
}

type withHTTP2 bool

func (w withHTTP2) apply(o *options) {
	o.http2 = bool(w)
}

func WithHTTP2(enable bool) TransportOption {
	return withHTTP2(enable)
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

	if t.opts.http2 {
		if t.opts.insecure {
			t.Conn, err = http2TransportInsecure(address)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if t.opts.insecure {
			t.Conn, err = quicTransport(address)
			if err != nil {
				return nil, err
			}
		}
	}

	if t.Conn == nil {
		return nil, fmt.Errorf("couldn't initialize grpc client")
	}

	return t.Conn, nil
}

func http2TransportInsecure(address string) (*grpc.ClientConn, error) {
	target := fmt.Sprintf("%s:%d", address, 1337)
	return grpc.Dial(target, grpc.WithInsecure())
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
