package certlib

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"git.wntrmute.dev/kyle/goutils/certlib/certerr"
)

// ReadCertificate reads a DER or PEM-encoded certificate from the
// byte slice.
func ReadCertificate(in []byte) (cert *x509.Certificate, rest []byte, err error) {
	if len(in) == 0 {
		err = certerr.ErrEmptyCertificate
		return cert, rest, err
	}

	if in[0] == '-' {
		p, remaining := pem.Decode(in)
		if p == nil {
			err = errors.New("certlib: invalid PEM file")
			return cert, rest, err
		}

		rest = remaining
		if p.Type != "CERTIFICATE" {
			err = certerr.ErrInvalidPEMType(p.Type, "CERTIFICATE")
			return cert, rest, err
		}

		in = p.Bytes
	}

	cert, err = x509.ParseCertificate(in)
	return cert, rest, err
}

// ReadCertificates tries to read all the certificates in a
// PEM-encoded collection.
func ReadCertificates(in []byte) (certs []*x509.Certificate, err error) {
	var cert *x509.Certificate
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
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cert, _, err := ReadCertificate(in)
	return cert, err
}

// LoadCertificates tries to read all the certificates in a file,
// returning them in the order that it found them in the file.
func LoadCertificates(path string) ([]*x509.Certificate, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ReadCertificates(in)
}
