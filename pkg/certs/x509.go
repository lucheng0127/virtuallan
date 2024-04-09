package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
)

func NewClientTlsConf() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic"},
	}
}

func NewServerTlsConf() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDer, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	crtPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})

	tlsCert, err := tls.X509KeyPair(crtPem, keyPem)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic"},
	}, nil
}
