package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

const (
	keyTypeRSA   = "RSA"
	keyTypeECDSA = "ECDSA"
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `ski: print subject key info for PEM-encoded files

Usage:
	ski [-hm] files...

Flags:
	-d  Hex encoding mode.
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

func dumpHex(in []byte, mode lib.HexEncodeMode) string {
	return lib.HexEncode(in, mode)
}

type subjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

func main() {
	var help, shouldMatch bool
	var displayModeString string
	flag.StringVar(&displayModeString, "d", "lower", "hex encoding mode")
	flag.BoolVar(&help, "h", false, "print a help message and exit")
	flag.BoolVar(&shouldMatch, "m", false, "all SKIs should match")
	flag.Parse()

	displayMode := lib.ParseHexEncodeMode(displayModeString)

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

		pubHash := sha1.Sum(subPKI.SubjectPublicKey.Bytes) // #nosec G401 this is the standard
		pubHashString := dumpHex(pubHash[:], displayMode)
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
