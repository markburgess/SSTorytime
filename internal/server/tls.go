package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// EnsureTLSMaterial loads cert/key PEMs from disk, or generates a self-signed
// development certificate with crypto/* (stdlib) if either file is missing.
// Matches upstream intent (local self-signed TLS) without requiring openssl.
func EnsureTLSMaterial(certPath, keyPath string) (certFile, keyFile string, generated bool, err error) {
	if certPath == "" {
		certPath = "cert.pem"
	}
	if keyPath == "" {
		keyPath = "key.pem"
	}
	_, cerr := os.Stat(certPath)
	_, kerr := os.Stat(keyPath)
	if cerr == nil && kerr == nil {
		return certPath, keyPath, false, nil
	}
	if err := writeSelfSigned(certPath, keyPath); err != nil {
		return "", "", false, err
	}
	return certPath, keyPath, true, nil
}

func writeSelfSigned(certPath, keyPath string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("tls key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("tls serial: %w", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"SSTorytime"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("tls cert: %w", err)
	}

	certOut, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("tls cert write: %w", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return fmt.Errorf("tls cert pem: %w", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("tls key marshal: %w", err)
	}
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("tls key write: %w", err)
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return fmt.Errorf("tls key pem: %w", err)
	}
	return nil
}

// ErrMissingTLSHalf is returned when exactly one of cert/key exists.
var ErrMissingTLSHalf = errors.New("tls: cert and key must both exist or both be absent (will auto-generate)")
