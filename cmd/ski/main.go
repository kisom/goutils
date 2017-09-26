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
	"io/ioutil"
	"os"
	"strings"

	"github.com/kisom/goutils/die"
	"github.com/kisom/goutils/lib"
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

func parse(path string) (public []byte, kt, ft string) {
	data, err := ioutil.ReadFile(path)
	die.If(err)

	data = bytes.TrimSpace(data)
	p, rest := pem.Decode(data)
	if len(rest) > 0 {
		lib.Warnx("trailing data in PEM file")
	}

	data = p.Bytes

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

	return
}

func parseKey(data []byte) (public []byte, kt string) {
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
	switch privInterface.(type) {
	case *rsa.PrivateKey:
		priv = privInterface.(*rsa.PrivateKey)
		kt = "RSA"
	case *ecdsa.PrivateKey:
		priv = privInterface.(*ecdsa.PrivateKey)
		kt = "ECDSA"
	default:
		die.With("unknown private key type %T", privInterface)
	}

	public, err = x509.MarshalPKIXPublicKey(priv.Public())
	die.If(err)

	return
}

func parseCertificate(data []byte) (public []byte, kt string) {
	cert, err := x509.ParseCertificate(data)
	die.If(err)

	pub := cert.PublicKey
	switch pub.(type) {
	case *rsa.PublicKey:
		kt = "RSA"
	case *ecdsa.PublicKey:
		kt = "ECDSA"
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err = x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return
}

func parseCSR(data []byte) (public []byte, kt string) {
	csr, err := x509.ParseCertificateRequest(data)
	die.If(err)

	pub := csr.PublicKey
	switch pub.(type) {
	case *rsa.PublicKey:
		kt = "RSA"
	case *ecdsa.PublicKey:
		kt = "ECDSA"
	default:
		die.With("unknown public key type %T", pub)
	}

	public, err = x509.MarshalPKIXPublicKey(pub)
	die.If(err)
	return
}

func dumpHex(in []byte) string {
	var s string
	for i := range in {
		s += fmt.Sprintf("%02X:", in[i])
	}

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
			lib.Warn(err, "failed to get subject PKI")
			continue
		}

		pubHash := sha1.Sum(subPKI.SubjectPublicKey.Bytes)
		pubHashString := dumpHex(pubHash[:])
		if ski == "" {
			ski = pubHashString
		}

		if shouldMatch && ski != pubHashString {
			lib.Warnx("%s: SKI mismatch (%s != %s)",
				path, ski, pubHashString)
		}
		fmt.Printf("%s  %s (%s %s)\n", path, pubHashString, kt, ft)
	}
}
