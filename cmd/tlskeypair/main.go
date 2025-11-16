package main

import (
    "bytes"
    "crypto"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "errors"
    "flag"
    "fmt"
    "os"

    "git.wntrmute.dev/kyle/goutils/die"
)

var validPEMs = map[string]bool{
	"PRIVATE KEY":     true,
	"RSA PRIVATE KEY": true,
	"EC PRIVATE KEY":  true,
}

const (
	curveInvalid = iota // any invalid curve
	curveRSA            // indicates key is an RSA key, not an EC key
	curveP256
	curveP384
	curveP521
)

func getECCurve(pub interface{}) int {
	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P256():
			return curveP256
		case elliptic.P384():
			return curveP384
		case elliptic.P521():
			return curveP521
		default:
			return curveInvalid
		}
	case *rsa.PublicKey:
		return curveRSA
	default:
		return curveInvalid
	}
}

// matchRSA compares an RSA public key from certificate against RSA public key from private key.
// It returns true on match.
func matchRSA(certPub *rsa.PublicKey, keyPub *rsa.PublicKey) bool {
    return keyPub.N.Cmp(certPub.N) == 0 && keyPub.E == certPub.E
}

// matchECDSA compares ECDSA public keys for equality and compatible curve.
// It returns match=true when they are on the same curve and have the same X/Y.
// If curves mismatch, match is false.
func matchECDSA(certPub *ecdsa.PublicKey, keyPub *ecdsa.PublicKey) bool {
    if getECCurve(certPub) != getECCurve(keyPub) {
        return false
    }
    if keyPub.X.Cmp(certPub.X) != 0 {
        return false
    }
    if keyPub.Y.Cmp(certPub.Y) != 0 {
        return false
    }
    return true
}

// matchKeys determines whether the certificate's public key matches the given private key.
// It returns true if they match; otherwise, it returns false and a human-friendly reason.
func matchKeys(cert *x509.Certificate, priv crypto.Signer) (bool, string) {
    switch keyPub := priv.Public().(type) {
    case *rsa.PublicKey:
        switch certPub := cert.PublicKey.(type) {
        case *rsa.PublicKey:
            if matchRSA(certPub, keyPub) {
                return true, ""
            }
            return false, "public keys don't match"
        case *ecdsa.PublicKey:
            return false, "RSA private key, EC public key"
        default:
            return false, fmt.Sprintf("unsupported certificate public key type: %T", cert.PublicKey)
        }
    case *ecdsa.PublicKey:
        switch certPub := cert.PublicKey.(type) {
        case *ecdsa.PublicKey:
            if matchECDSA(certPub, keyPub) {
                return true, ""
            }
            // Determine a more precise reason
            kc := getECCurve(keyPub)
            cc := getECCurve(certPub)
            if kc == curveInvalid {
                return false, "invalid private key curve"
            }
            if cc == curveRSA {
                return false, "private key is EC, certificate is RSA"
            }
            if kc != cc {
                return false, "EC curves don't match"
            }
            return false, "public keys don't match"
        case *rsa.PublicKey:
            return false, "private key is EC, certificate is RSA"
        default:
            return false, fmt.Sprintf("unsupported certificate public key type: %T", cert.PublicKey)
        }
    default:
        return false, fmt.Sprintf("unrecognised private key type: %T", priv.Public())
    }
}

func loadKey(path string) (crypto.Signer, error) {
    in, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	in = bytes.TrimSpace(in)
	p, _ := pem.Decode(in)
	if p != nil {
		if !validPEMs[p.Type] {
			return nil, errors.New("invalid private key file type " + p.Type)
		}
		in = p.Bytes
	}

 priv, err := x509.ParsePKCS8PrivateKey(in)
	if err != nil {
		priv, err = x509.ParsePKCS1PrivateKey(in)
		if err != nil {
			priv, err = x509.ParseECPrivateKey(in)
			if err != nil {
				return nil, err
			}
		}
	}

 switch p := priv.(type) {
 case *rsa.PrivateKey:
     return p, nil
 case *ecdsa.PrivateKey:
     return p, nil
 default:
     // should never reach here
     return nil, errors.New("invalid private key")
 }

}

func main() {
	var keyFile, certFile string
	flag.StringVar(&keyFile, "k", "", "TLS private `key` file")
	flag.StringVar(&certFile, "c", "", "TLS `certificate` file")
	flag.Parse()

 in, err := os.ReadFile(certFile)
	die.If(err)

	p, _ := pem.Decode(in)
	if p != nil {
		if p.Type != "CERTIFICATE" {
			die.With("invalid certificate (type is %s)", p.Type)
		}
		in = p.Bytes
	}
	cert, err := x509.ParseCertificate(in)
	die.If(err)

	priv, err := loadKey(keyFile)
	die.If(err)

 matched, reason := matchKeys(cert, priv)
 if matched {
     fmt.Println("Match.")
     return
 }
 fmt.Printf("No match (%s).\n", reason)
 os.Exit(1)
}
