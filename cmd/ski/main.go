package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `ski: print subject key info for PEM-encoded files

Usage:
	ski [-hm] files...

Flags:
	-h	Print this help message.
	-m	All SKIs should match; as soon as an SKI mismatch is found,
		it is reported.

`)
}

func init() {
	flag.Usage = func() { usage(os.Stderr) }
}

func parse(path string) ([]byte, string, string) {
	data, err := os.ReadFile(path)
	die.If(err)

	data = bytes.TrimSpace(data)
	p, rest := pem.Decode(data)
	if len(rest) > 0 {
		_, _ = lib.Warnx("trailing data in PEM file")
	}

	if p == nil {
		die.With("no PEM data found")
	}

	data = p.Bytes

	var (
		public []byte
		kt     string
		ft     string
	)

	switch p.Type {
	case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
		public, kt = parseKey(data)
		ft = "private key"
	case "CERTIFICATE":
		public, kt = parseCertificate(data)
		ft = "certificate"
	case "CERTIFICATE REQUEST":
		public, kt = parseCSR(data)
		ft = "certificate request"
	default:
		die.With("unknown PEM type %s", p.Type)
	}

	return public, kt, ft
}

func parseKey(data []byte) ([]byte, string) {
	privInterface, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		privInterface, err = x509.ParsePKCS1PrivateKey(data)
		if err != nil {
			privInterface, err = x509.ParseECPrivateKey(data)
			if err != nil {
				die.With("couldn't parse private key.")
			}
		}
	}

	var priv crypto.Signer
	var kt string
	switch p := privInterface.(type) {
	case *rsa.PrivateKey:
		priv = p
		kt = "RSA"
	case *ecdsa.PrivateKey:
		priv = p
		kt = "ECDSA"
	default:
		die.With("unknown private key type %T", privInterface)
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
		kt = "RSA"
	case *ecdsa.PublicKey:
		kt = "ECDSA"
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err := x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return public, kt
}

func parseCSR(data []byte) ([]byte, string) {
	csr, err := x509.ParseCertificateRequest(data)
	die.If(err)

	pub := csr.PublicKey
	var kt string
	switch pub.(type) {
	case *rsa.PublicKey:
		kt = "RSA"
	case *ecdsa.PublicKey:
		kt = "ECDSA"
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err := x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return public, kt
}

func dumpHex(in []byte) string {
	var s string
	var sSb153 strings.Builder
	for i := range in {
		sSb153.WriteString(fmt.Sprintf("%02X:", in[i]))
	}
	s += sSb153.String()

	return strings.Trim(s, ":")
}

type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

func main() {
	var help, shouldMatch bool
	flag.BoolVar(&help, "h", false, "print a help message and exit")
	flag.BoolVar(&shouldMatch, "m", false, "all SKIs should match")
	flag.Parse()

	if help {
		usage(os.Stdout)
		os.Exit(0)
	}

	var ski string
	for _, path := range flag.Args() {
		public, kt, ft := parse(path)

		var subPKI subjectPublicKeyInfo
		_, err := asn1.Unmarshal(public, &subPKI)
		if err != nil {
			_, _ = lib.Warn(err, "failed to get subject PKI")
			continue
		}

		pubHash := sha1.Sum(subPKI.SubjectPublicKey.Bytes)
		pubHashString := dumpHex(pubHash[:])
		if ski == "" {
			ski = pubHashString
		}

		if shouldMatch && ski != pubHashString {
			_, _ = lib.Warnx("%s: SKI mismatch (%s != %s)",
				path, ski, pubHashString)
		}
		fmt.Printf("%s  %s (%s %s)\n", path, pubHashString, kt, ft)
	}
}
