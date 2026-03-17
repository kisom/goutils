package certgen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"time"
)

// TestCA is an in-memory certificate authority for use in tests. It
// provides a root CA certificate and the ability to issue leaf
// certificates for TLS testing with full verification enabled.
type TestCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

// NewTestCA creates a new TestCA with a self-signed P-256 root
// certificate. The CA is valid for 1 hour.
func NewTestCA() (*TestCA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("certgen: generating CA key: %w", err)
	}

	serial, err := SerialNumber()
	if err != nil {
		return nil, fmt.Errorf("certgen: generating serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "Test CA",
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("certgen: creating CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("certgen: parsing CA certificate: %w", err)
	}

	return &TestCA{cert: cert, key: key}, nil
}

// Certificate returns the root CA certificate.
func (ca *TestCA) Certificate() *x509.Certificate {
	return ca.cert
}

// CertificatePEM returns the root CA certificate as a PEM-encoded
// byte slice.
func (ca *TestCA) CertificatePEM() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.cert.Raw,
	})
}

// CertPool returns a certificate pool containing the root CA
// certificate, suitable for use as a TLS root CA pool.
func (ca *TestCA) CertPool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)
	return pool
}

// Issue creates a new leaf certificate signed by the CA for the given
// DNS names and IP addresses. It returns the leaf private key and
// certificate. The leaf certificate is valid for 1 hour with key
// usage appropriate for a TLS server.
func (ca *TestCA) Issue(dnsNames []string, ips []net.IP) (crypto.Signer, *x509.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("certgen: generating leaf key: %w", err)
	}

	serial, err := SerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("certgen: generating serial: %w", err)
	}

	cn := "localhost"
	if len(dnsNames) > 0 {
		cn = dnsNames[0]
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now().Add(-1 * time.Minute),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
		IPAddresses:           ips,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return nil, nil, fmt.Errorf("certgen: creating leaf certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("certgen: parsing leaf certificate: %w", err)
	}

	return key, cert, nil
}

// IssueServer is a convenience wrapper around Issue for the common
// case of a server certificate for localhost (both DNS and IP).
func (ca *TestCA) IssueServer() (crypto.Signer, *x509.Certificate, error) {
	return ca.Issue(
		[]string{"localhost"},
		[]net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	)
}

// TLSConfig returns a tls.Config (by value) configured with the CA's
// root pool for verification. The caller can set additional fields
// (e.g., Certificates) or modify the returned config safely.
func (ca *TestCA) TLSConfig() tls.Config {
	return tls.Config{
		RootCAs:    ca.CertPool(),
		MinVersion: tls.VersionTLS13,
	}
}

// ServerTLSConfig returns a tls.Config (by value) for a TLS server
// using the given leaf key and certificate, with client verification
// against the CA root pool. Pass key and cert from Issue or
// IssueServer.
func (ca *TestCA) ServerTLSConfig(key crypto.Signer, cert *x509.Certificate) tls.Config {
	return tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  key,
				Leaf:        cert,
			},
		},
		ClientCAs:  ca.CertPool(),
		MinVersion: tls.VersionTLS13,
	}
}

// TLSKeyPair returns a tls.Certificate from the given key and
// certificate, suitable for use in a tls.Config.Certificates slice.
func TLSKeyPair(key crypto.Signer, cert *x509.Certificate) tls.Certificate {
	return tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  key,
		Leaf:        cert,
	}
}

// MustTestCA calls NewTestCA and panics on error. Intended for use
// in TestMain or test helpers where error handling is impractical.
func MustTestCA() *TestCA {
	ca, err := NewTestCA()
	if err != nil {
		panic("certgen: " + err.Error())
	}
	return ca
}

// CertificatePEM returns a PEM-encoded byte slice for the given
// certificate.
func CertificatePEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
}

// PrivateKeyPEM returns a PEM-encoded PKCS#8 byte slice for the
// given private key.
func PrivateKeyPEM(key crypto.Signer) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("certgen: marshaling private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}), nil
}

