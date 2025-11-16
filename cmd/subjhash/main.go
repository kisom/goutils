package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func init() {
	flag.Usage = func() { usage(os.Stdout); os.Exit(1) }
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `Print hash of subject or issuer fields in certificates.

Usage: subjhash [-im] certs...

Flags:
	-i	Print hash of issuer field.
	-m	Matching mode. This expects arguments to be in the form of
		pairs of certificates (e.g. previous, new) whose subjects
		will be compared. For example,

			subjhash -m ca1.pem ca1-renewed.pem	\
				ca2.pem ca2-renewed.pem

		will exit with a non-zero status if the subject in the
		ca1-renewed.pem certificate doesn't match the subject in the
		ca.pem certificate; similarly for ca2.
`)
}

// NB: the Issuer field is *also* a subject field. Also, the returned
// hash is *not* hex encoded.
func getSubjectInfoHash(cert *x509.Certificate, issuer bool) []byte {
	if cert == nil {
		return nil
	}

	var subject []byte
	if issuer {
		subject = cert.RawIssuer
	} else {
		subject = cert.RawSubject
	}

	digest := sha256.Sum256(subject)
	return digest[:]
}

func printDigests(paths []string, issuer bool) {
	for _, path := range paths {
		cert, err := certlib.LoadCertificate(path)
		if err != nil {
			_, _ = lib.Warn(err, "failed to load certificate from %s", path)
			continue
		}

		digest := getSubjectInfoHash(cert, issuer)
		fmt.Printf("%x  %s\n", digest, path)
	}
}

func matchDigests(paths []string, issuer bool) {
	if (len(paths) % 2) != 0 {
		lib.Errx(lib.ExitFailure, "not all certificates are paired")
	}

	var invalid int
	for len(paths) > 0 {
		fst := paths[0]
		snd := paths[1]
		paths = paths[2:]

		fstCert, err := certlib.LoadCertificate(fst)
		die.If(err)

		sndCert, err := certlib.LoadCertificate(snd)
		die.If(err)

		if !bytes.Equal(getSubjectInfoHash(fstCert, issuer), getSubjectInfoHash(sndCert, issuer)) {
			_, _ = lib.Warnx("certificates don't match: %s and %s", fst, snd)
			invalid++
		}
	}

	if invalid > 0 {
		os.Exit(1)
	}
}

func main() {
	var issuer, match bool
	flag.BoolVar(&issuer, "i", false, "print the issuer")
	flag.BoolVar(&match, "m", false, "match mode")
	flag.Parse()

	paths := flag.Args()
	if match {
		matchDigests(paths, issuer)
	} else {
		printDigests(paths, issuer)
	}
}
