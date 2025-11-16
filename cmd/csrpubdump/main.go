package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
)

func main() {
	flag.Parse()

	for _, fileName := range flag.Args() {
		in, err := os.ReadFile(fileName)
		die.If(err)

		csr, _, err := certlib.ParseCSR(in)
		die.If(err)

		out, err := x509.MarshalPKIXPublicKey(csr.PublicKey)
		die.If(err)

		var t string
		switch pub := csr.PublicKey.(type) {
		case *rsa.PublicKey:
			t = "RSA PUBLIC KEY"
		case *ecdsa.PublicKey:
			t = "EC PUBLIC KEY"
		default:
			die.With("unrecognised public key type %T", pub)
		}

		p := &pem.Block{
			Type:  t,
			Bytes: out,
		}

		err = os.WriteFile(fileName+".pub", pem.EncodeToMemory(p), 0o644) // #nosec G306
		die.If(err)
		fmt.Fprintf(os.Stdout, "[+] wrote %s.\n", fileName+".pub")
	}
}
