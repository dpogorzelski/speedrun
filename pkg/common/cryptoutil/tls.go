package cryptoutil

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"math/big"
)

func InsecureTLSConfig() (*tls.Config, error) {
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
		MinVersion:         tls.VersionTLS13,
		CurvePreferences:   []tls.CurveID{tls.X25519},
		CipherSuites:       []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true,
	}, nil
}

func ClientTLSConfig(caPath, certPath, keyPath string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519},
		CipherSuites:     []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		Certificates:     []tls.Certificate{cert},
		RootCAs:          caCertPool,
	}, nil
}

func ServerTLSConfig(caPath, certPath, keyPath string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.X25519},
		CipherSuites:     []uint16{tls.TLS_CHACHA20_POLY1305_SHA256},
		Certificates:     []tls.Certificate{cert},
		ClientCAs:        caCertPool,
		ClientAuth:       tls.RequireAndVerifyClientCert,
	}, nil
}
