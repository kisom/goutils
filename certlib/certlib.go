package certlib

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib/certerr"
)

// ReadCertificate reads a DER or PEM-encoded certificate from the
// byte slice.
func ReadCertificate(in []byte) (*x509.Certificate, []byte, error) {
	in = bytes.TrimSpace(in)
	if len(in) == 0 {
		return nil, nil, certerr.ParsingError(certerr.ErrorSourceCertificate, certerr.ErrEmptyCertificate)
	}

	if in[0] == '-' {
		p, remaining := pem.Decode(in)
		if p == nil {
			return nil, nil, certerr.ParsingError(certerr.ErrorSourceCertificate, errors.New("invalid PEM file"))
		}

		rest := remaining
		if p.Type != pemTypeCertificate {
			return nil, rest, certerr.ParsingError(
				certerr.ErrorSourceCertificate,
				certerr.ErrInvalidPEMType(p.Type, pemTypeCertificate),
			)
		}

		in = p.Bytes
		cert, err := x509.ParseCertificate(in)
		if err != nil {
			return nil, rest, certerr.ParsingError(certerr.ErrorSourceCertificate, err)
		}
		return cert, rest, nil
	}

	cert, err := x509.ParseCertificate(in)
	if err != nil {
		return nil, nil, certerr.ParsingError(certerr.ErrorSourceCertificate, err)
	}
	return cert, nil, nil
}

// ReadCertificates tries to read all the certificates in a
// PEM-encoded collection.
func ReadCertificates(in []byte) ([]*x509.Certificate, error) {
	var cert *x509.Certificate
	var certs []*x509.Certificate
	var err error
	for {
		cert, in, err = ReadCertificate(in)
		if err != nil {
			break
		}

		if cert == nil {
			break
		}

		certs = append(certs, cert)
		if len(in) == 0 {
			break
		}
	}

	return certs, err
}

// LoadCertificate tries to read a single certificate from disk. If
// the file contains multiple certificates (e.g. a chain), only the
// first certificate is returned.
func LoadCertificate(path string) (*x509.Certificate, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, certerr.LoadingError(certerr.ErrorSourceCertificate, err)
	}

	cert, _, err := ReadCertificate(in)
	return cert, err
}

// LoadCertificates tries to read all the certificates in a file,
// returning them in the order that it found them in the file.
func LoadCertificates(path string) ([]*x509.Certificate, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, certerr.LoadingError(certerr.ErrorSourceCertificate, err)
	}

	return ReadCertificates(in)
}

func PoolFromBytes(certBytes []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	certs, err := ReadCertificates(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificates: %w", err)
	}

	for _, cert := range certs {
		pool.AddCert(cert)
	}

	return pool, nil
}

func ExportPrivateKeyPEM(priv crypto.PrivateKey) ([]byte, error) {
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: pemTypePrivateKey, Bytes: keyDER}), nil
}
