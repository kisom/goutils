package certlib

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

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"

	"git.wntrmute.dev/kyle/goutils/certlib/certerr"
)

// ParsePrivateKeyDER parses a PKCS #1, PKCS #8, ECDSA, or Ed25519 DER-encoded
// private key. The key must not be in PEM format. If an error is returned, it
// may contain information about the private key, so care should be taken when
// displaying it directly.
func ParsePrivateKeyDER(keyDER []byte) (crypto.Signer, error) {
	// Try common encodings in order without deep nesting.
	if k, err := x509.ParsePKCS8PrivateKey(keyDER); err == nil {
		switch kk := k.(type) {
		case *rsa.PrivateKey:
			return kk, nil
		case *ecdsa.PrivateKey:
			return kk, nil
		case ed25519.PrivateKey:
			return kk, nil
		default:
			return nil, certerr.ParsingError(certerr.ErrorSourcePrivateKey, fmt.Errorf("unknown key type %T", k))
		}
	}
	if k, err := x509.ParsePKCS1PrivateKey(keyDER); err == nil {
		return k, nil
	}
	if k, err := x509.ParseECPrivateKey(keyDER); err == nil {
		return k, nil
	}
	if k, err := ParseEd25519PrivateKey(keyDER); err == nil {
		if kk, ok := k.(ed25519.PrivateKey); ok {
			return kk, nil
		}
		return nil, certerr.ParsingError(certerr.ErrorSourcePrivateKey, fmt.Errorf("unknown key type %T", k))
	}
	// If all parsers failed, return the last error from Ed25519 attempt (approximate cause).
	if _, err := ParseEd25519PrivateKey(keyDER); err != nil {
		return nil, certerr.ParsingError(certerr.ErrorSourcePrivateKey, err)
	}
	// Fallback (should be unreachable)
	return nil, certerr.ParsingError(certerr.ErrorSourcePrivateKey, errors.New("unknown key encoding"))
}
