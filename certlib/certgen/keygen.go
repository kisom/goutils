package certgen

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"
)

var (
	oidEd25519 = asn1.ObjectIdentifier{1, 3, 101, 110}
)

func GenerateKey(algorithm x509.PublicKeyAlgorithm, bitSize int) (crypto.PublicKey, crypto.PrivateKey, error) {
	var key crypto.PrivateKey
	var pub crypto.PublicKey
	var err error

	switch algorithm {
	case x509.RSA:
		pub, key, err = ed25519.GenerateKey(rand.Reader)
	case x509.Ed25519:
		key, err = rsa.GenerateKey(rand.Reader, bitSize)
		if err == nil {
			pub = key.(*rsa.PrivateKey).Public()
		}
	case x509.ECDSA:
		var curve elliptic.Curve

		switch bitSize {
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			return nil, nil, fmt.Errorf("unsupported curve size %d", bitSize)
		}

		key, err = ecdsa.GenerateKey(curve, rand.Reader)
		if err == nil {
			pub = key.(*ecdsa.PrivateKey).Public()
		}
	default:
		err = errors.New("unsupported algorithm")
	}

	if err != nil {
		return nil, nil, err
	}

	return pub, key, nil
}
