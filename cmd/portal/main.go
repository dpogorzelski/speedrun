package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/text"
	sysctl "github.com/lorenzosaino/go-sysctl"
	"github.com/lucas-clemente/quic-go"
	qnet "github.com/speedrunsh/grpc-quic"
	"github.com/speedrunsh/speedrun/pkg/portal"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"google.golang.org/grpc"
)

const addr = "0.0.0.0:1337"

func main() {
	h := loghandler.New(os.Stdout)
	log.SetHandler(h)
	err := sysctl.Set("net.core.rmem_max", "2500000")
	if err != nil {
		log.Fatalf("couldn't set net.core.rmem_max: %s", err.Error())
	}

	tlsConf, err := generateTLSConfig()
	if err != nil {
		log.Fatalf("QuicServer: failed to generateTLSConfig. %s", err.Error())
	}

	ql, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		log.Fatalf("QuicServer: failed to ListenAddr. %s", err.Error())
	}
	listener := qnet.Listen(ql)

	s := grpc.NewServer()
	portalpb.RegisterPortalServer(s, &portal.Server{})
	log.Infof("Started portal on %s", addr)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func generateTLSConfig() (*tls.Config, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return nil, err
	}
	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519},
		CipherSuites:     []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		Certificates:     []tls.Certificate{tlsCert},
		NextProtos:       []string{"speedrun"},
	}, nil
}
