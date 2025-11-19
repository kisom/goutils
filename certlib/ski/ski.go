package ski

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505 this is the standard
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

const (
	keyTypeRSA     = "RSA"
	keyTypeECDSA   = "ECDSA"
	keyTypeEd25519 = "Ed25519"
)

type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

type KeyInfo struct {
	PublicKey []byte
	KeyType   string
	FileType  string
}

func (k *KeyInfo) String() string {
	return fmt.Sprintf("%s (%s)", lib.HexEncode(k.PublicKey, lib.HexEncodeLowerColon), k.KeyType)
}

func (k *KeyInfo) SKI(displayMode lib.HexEncodeMode) (string, error) {
	var subPKI subjectPublicKeyInfo

	_, err := asn1.Unmarshal(k.PublicKey, &subPKI)
	if err != nil {
		return "", fmt.Errorf("serializing SKI: %w", err)
	}

	pubHash := sha1.Sum(subPKI.SubjectPublicKey.Bytes) // #nosec G401 this is the standard
	pubHashString := lib.HexEncode(pubHash[:], displayMode)

	return pubHashString, nil
}

// ParsePEM parses a PEM file and returns the public key and its type.
func ParsePEM(path string) (*KeyInfo, error) {
	material := &KeyInfo{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing X.509 material %s: %w", path, err)
	}

	data = bytes.TrimSpace(data)
	p, rest := pem.Decode(data)
	if len(rest) > 0 {
		lib.Warnx("trailing data in PEM file")
	}

	if p == nil {
		return nil, fmt.Errorf("no PEM data in %s", path)
	}

	data = p.Bytes

	switch p.Type {
	case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
		material.PublicKey, material.KeyType = parseKey(data)
		material.FileType = "private key"
	case "CERTIFICATE":
		material.PublicKey, material.KeyType = parseCertificate(data)
		material.FileType = "certificate"
	case "CERTIFICATE REQUEST":
		material.PublicKey, material.KeyType = parseCSR(data)
		material.FileType = "certificate request"
	default:
		return nil, fmt.Errorf("unknown PEM type %s", p.Type)
	}

	return material, nil
}

func parseKey(data []byte) ([]byte, string) {
	priv, err := certlib.ParsePrivateKeyDER(data)
	if err != nil {
		die.If(err)
	}

	var kt string
	switch priv.Public().(type) {
	case *rsa.PublicKey:
		kt = keyTypeRSA
	case *ecdsa.PublicKey:
		kt = keyTypeECDSA
	default:
		die.With("unknown private key type %T", priv)
	}

	public, err := x509.MarshalPKIXPublicKey(priv.Public())
	die.If(err)

	return public, kt
}

func parseCertificate(data []byte) ([]byte, string) {
	cert, err := x509.ParseCertificate(data)
	die.If(err)

	pub := cert.PublicKey
	var kt string
	switch pub.(type) {
	case *rsa.PublicKey:
		kt = keyTypeRSA
	case *ecdsa.PublicKey:
		kt = keyTypeECDSA
	case *ed25519.PublicKey:
		kt = keyTypeEd25519
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err := x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return public, kt
}

func parseCSR(data []byte) ([]byte, string) {
	// Use certlib to support both PEM and DER and to centralize validation.
	csr, _, err := certlib.ParseCSR(data)
	die.If(err)

	pub := csr.PublicKey
	var kt string
	switch pub.(type) {
	case *rsa.PublicKey:
		kt = keyTypeRSA
	case *ecdsa.PublicKey:
		kt = keyTypeECDSA
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err := x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return public, kt
}
