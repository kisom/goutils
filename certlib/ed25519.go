package certlib

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
)

// Originally from CFSSL, mostly written by me originally, and licensed under:

/*
Copyright (c) 2014 CloudFlare Inc.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// I've modified it for use in my own code e.g. by removing the CFSSL errors
// and replacing them with sane ones.

var errEd25519WrongID = errors.New("incorrect object identifier")
var errEd25519WrongKeyType = errors.New("incorrect key type")

// ed25519OID is the OID for the Ed25519 signature scheme: see
// https://datatracker.ietf.org/doc/draft-ietf-curdle-pkix-04.
var ed25519OID = asn1.ObjectIdentifier{1, 3, 101, 112}

// subjectPublicKeyInfo reflects the ASN.1 object defined in the X.509 standard.
//
// This is defined in crypto/x509 as "publicKeyInfo".
type subjectPublicKeyInfo struct {
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

// MarshalEd25519PublicKey creates a DER-encoded SubjectPublicKeyInfo for an
// ed25519 public key, as defined in
// https://tools.ietf.org/html/draft-ietf-curdle-pkix-04. This is analogous to
// MarshalPKIXPublicKey in crypto/x509, which doesn't currently support Ed25519.
func MarshalEd25519PublicKey(pk crypto.PublicKey) ([]byte, error) {
	pub, ok := pk.(ed25519.PublicKey)
	if !ok {
		return nil, errEd25519WrongKeyType
	}

	spki := subjectPublicKeyInfo{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: ed25519OID,
		},
		PublicKey: asn1.BitString{
			BitLength: len(pub) * 8,
			Bytes:     pub,
		},
	}

	return asn1.Marshal(spki)
}

// ParseEd25519PublicKey returns the Ed25519 public key encoded by the input.
func ParseEd25519PublicKey(der []byte) (crypto.PublicKey, error) {
	var spki subjectPublicKeyInfo
	if rest, err := asn1.Unmarshal(der, &spki); err != nil {
		return nil, err
	} else if len(rest) > 0 {
		return nil, errors.New("SubjectPublicKeyInfo too long")
	}

	if !spki.Algorithm.Algorithm.Equal(ed25519OID) {
		return nil, errEd25519WrongID
	}

	if spki.PublicKey.BitLength != ed25519.PublicKeySize*8 {
		return nil, errors.New("SubjectPublicKeyInfo PublicKey length mismatch")
	}

	return ed25519.PublicKey(spki.PublicKey.Bytes), nil
}

// oneAsymmetricKey reflects the ASN.1 structure for storing private keys in
// https://tools.ietf.org/html/draft-ietf-curdle-pkix-04, excluding the optional
// fields, which we don't use here.
//
// This is identical to pkcs8 in crypto/x509.
type oneAsymmetricKey struct {
	Version    int
	Algorithm  pkix.AlgorithmIdentifier
	PrivateKey []byte
}

// curvePrivateKey is the innter type of the PrivateKey field of
// oneAsymmetricKey.
type curvePrivateKey []byte

// MarshalEd25519PrivateKey returns a DER encoding of the input private key as
// specified in https://tools.ietf.org/html/draft-ietf-curdle-pkix-04.
func MarshalEd25519PrivateKey(sk crypto.PrivateKey) ([]byte, error) {
	priv, ok := sk.(ed25519.PrivateKey)
	if !ok {
		return nil, errEd25519WrongKeyType
	}

	// Marshal the innter CurvePrivateKey.
	curvePrivateKey, err := asn1.Marshal(priv.Seed())
	if err != nil {
		return nil, err
	}

	// Marshal the OneAsymmetricKey.
	asym := oneAsymmetricKey{
		Version: 0,
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: ed25519OID,
		},
		PrivateKey: curvePrivateKey,
	}
	return asn1.Marshal(asym)
}

// ParseEd25519PrivateKey returns the Ed25519 private key encoded by the input.
func ParseEd25519PrivateKey(der []byte) (crypto.PrivateKey, error) {
	asym := new(oneAsymmetricKey)
	if rest, err := asn1.Unmarshal(der, asym); err != nil {
		return nil, err
	} else if len(rest) > 0 {
		return nil, errors.New("OneAsymmetricKey too long")
	}

	// Check that the key type is correct.
	if !asym.Algorithm.Algorithm.Equal(ed25519OID) {
		return nil, errEd25519WrongID
	}

	// Unmarshal the inner CurvePrivateKey.
	seed := new(curvePrivateKey)
	if rest, err := asn1.Unmarshal(asym.PrivateKey, seed); err != nil {
		return nil, err
	} else if len(rest) > 0 {
		return nil, errors.New("CurvePrivateKey too long")
	}

	return ed25519.NewKeyFromSeed(*seed), nil
}
