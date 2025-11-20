package certlib

import (
	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

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

func LoadCSR(path string) (*x509.CertificateRequest, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, certerr.LoadingError(certerr.ErrorSourceCSR, err)
	}

	req, _, err := ParseCSR(in)
	return req, err
}

func ExportCSRAsPEM(req *x509.CertificateRequest) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificateRequest, Bytes: req.Raw})
}

type FileFormat uint8

const (
	FormatPEM FileFormat = iota + 1
	FormatDER
)

func (f FileFormat) String() string {
	switch f {
	case FormatPEM:
		return "PEM"
	case FormatDER:
		return "DER"
	default:
		return "unknown"
	}
}

type KeyAlgo struct {
	Type  x509.PublicKeyAlgorithm
	Size  int
	curve elliptic.Curve
}

func (ka KeyAlgo) String() string {
	switch ka.Type {
	case x509.RSA:
		return fmt.Sprintf("RSA-%d", ka.Size)
	case x509.ECDSA:
		if ka.curve == nil {
			return fmt.Sprintf("ECDSA (unknown %d)", ka.Size)
		}
		return fmt.Sprintf("ECDSA-%s", ka.curve.Params().Name)
	case x509.Ed25519:
		return "Ed25519"
	case x509.DSA:
		return "DSA"
	case x509.UnknownPublicKeyAlgorithm:
		fallthrough // make linter happy
	default:
		return "unknown"
	}
}

func publicKeyAlgoFromPublicKey(key crypto.PublicKey) KeyAlgo {
	switch key := key.(type) {
	case *rsa.PublicKey:
		return KeyAlgo{
			Type: x509.RSA,
			Size: key.N.BitLen(),
		}
	case *ecdsa.PublicKey:
		return KeyAlgo{
			Type:  x509.ECDSA,
			curve: key.Curve,
			Size:  key.Params().BitSize,
		}
	case *ed25519.PublicKey:
		return KeyAlgo{
			Type: x509.Ed25519,
		}
	case *dsa.PublicKey:
		return KeyAlgo{
			Type: x509.DSA,
		}
	default:
		return KeyAlgo{
			Type: x509.UnknownPublicKeyAlgorithm,
		}
	}
}

func publicKeyAlgoFromKey(key crypto.PrivateKey) KeyAlgo {
	switch key := key.(type) {
	case *rsa.PrivateKey:
		return KeyAlgo{
			Type: x509.RSA,
			Size: key.PublicKey.N.BitLen(),
		}
	case *ecdsa.PrivateKey:
		return KeyAlgo{
			Type:  x509.ECDSA,
			curve: key.PublicKey.Curve,
			Size:  key.Params().BitSize,
		}
	case *ed25519.PrivateKey:
		return KeyAlgo{
			Type: x509.Ed25519,
		}
	case *dsa.PrivateKey:
		return KeyAlgo{
			Type: x509.DSA,
		}
	default:
		return KeyAlgo{
			Type: x509.UnknownPublicKeyAlgorithm,
		}
	}
}

func publicKeyAlgoFromCert(cert *x509.Certificate) KeyAlgo {
	return publicKeyAlgoFromPublicKey(cert.PublicKey)
}

func publicKeyAlgoFromCSR(csr *x509.CertificateRequest) KeyAlgo {
	return publicKeyAlgoFromPublicKey(csr.PublicKey)
}

type FileType struct {
	Format FileFormat
	Type   string
	Algo   KeyAlgo
}

func (ft FileType) String() string {
	if ft.Type == "" {
		return ft.Format.String()
	}
	return fmt.Sprintf("%s %s (%s)", ft.Algo, ft.Type, ft.Format)
}

// FileKind returns the file type of the given file.
func FileKind(path string) (*FileType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ft := &FileType{Format: FormatDER}

	block, _ := pem.Decode(data)
	if block != nil {
		data = block.Bytes
		ft.Type = strings.ToLower(block.Type)
		ft.Format = FormatPEM
	}

	cert, err := x509.ParseCertificate(data)
	if err == nil {
		ft.Algo = publicKeyAlgoFromCert(cert)
		return ft, nil
	}

	csr, err := x509.ParseCertificateRequest(data)
	if err == nil {
		ft.Algo = publicKeyAlgoFromCSR(csr)
		return ft, nil
	}

	priv, err := x509.ParsePKCS8PrivateKey(data)
	if err == nil {
		ft.Algo = publicKeyAlgoFromKey(priv)
		return ft, nil
	}

	return nil, errors.New("certlib; unknown file type")
}
