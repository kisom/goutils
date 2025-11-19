package main

import (
	"bytes"
	"crypto/x509"
	"embed"
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/certlib/verify"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

//go:embed testdata/*.pem
var embeddedTestdata embed.FS

type testCase struct {
	name     string
	caFile   string
	certFile string
	expectOK bool
}

func (tc testCase) Run() error {
	caBytes, err := embeddedTestdata.ReadFile(tc.caFile)
	if err != nil {
		return fmt.Errorf("selftest: failed to read embedded %s: %w", tc.caFile, err)
	}

	certBytes, err := embeddedTestdata.ReadFile(tc.certFile)
	if err != nil {
		return fmt.Errorf("selftest: failed to read embedded %s: %w", tc.certFile, err)
	}

	pool, err := certlib.PoolFromBytes(caBytes)
	if err != nil || pool == nil {
		return fmt.Errorf("selftest: failed to build CA pool for %s: %w", tc.caFile, err)
	}

	cert, _, err := certlib.ReadCertificate(certBytes)
	if err != nil {
		return fmt.Errorf("selftest: failed to parse certificate from %s: %w", tc.certFile, err)
	}

	_, err = verify.CertWith(cert, pool, nil, false)
	ok := err == nil

	if ok != tc.expectOK {
		return fmt.Errorf("%s: unexpected result: got %v, want %v", tc.name, ok, tc.expectOK)
	}

	if ok {
		fmt.Printf("%s: OK (expires %s)\n", tc.name, cert.NotAfter.Format(lib.DateShortFormat))
	}

	fmt.Printf("%s: INVALID (as expected)\n", tc.name)

	return nil
}

var cases = []testCase{
	{
		name:     "ISRG Root X1 validates LE E7",
		caFile:   "testdata/isrg-root-x1.pem",
		certFile: "testdata/le-e7.pem",
		expectOK: true,
	},
	{
		name:     "ISRG Root X1 does NOT validate Google WR2",
		caFile:   "testdata/isrg-root-x1.pem",
		certFile: "testdata/goog-wr2.pem",
		expectOK: false,
	},
	{
		name:     "GTS R1 validates Google WR2",
		caFile:   "testdata/gts-r1.pem",
		certFile: "testdata/goog-wr2.pem",
		expectOK: true,
	},
	{
		name:     "GTS R1 does NOT validate LE E7",
		caFile:   "testdata/gts-r1.pem",
		certFile: "testdata/le-e7.pem",
		expectOK: false,
	},
}

// selftest runs built-in validation using embedded certificates.
func selftest() int {
	failures := 0
	for _, tc := range cases {
		err := tc.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			failures++
			continue
		}
	}

	// Verify that both embedded root CAs are detected as self-signed
	roots := []string{"testdata/gts-r1.pem", "testdata/isrg-root-x1.pem"}
	for _, root := range roots {
		b, err := embeddedTestdata.ReadFile(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selftest: failed to read embedded %s: %v\n", root, err)
			failures++
			continue
		}

		certs, err := certlib.ReadCertificates(b)
		if err != nil || len(certs) == 0 {
			fmt.Fprintf(os.Stderr, "selftest: failed to parse cert(s) from %s: %v\n", root, err)
			failures++
			continue
		}

		leaf := certs[0]
		if len(leaf.AuthorityKeyId) == 0 || bytes.Equal(leaf.AuthorityKeyId, leaf.SubjectKeyId) {
			fmt.Printf("%s: SELF-SIGNED (as expected)\n", root)
		} else {
			fmt.Printf("%s: expected SELF-SIGNED, but was not detected as such\n", root)
			failures++
		}
	}

	if failures == 0 {
		fmt.Println("selftest: PASS")
		return 0
	}
	fmt.Fprintf(os.Stderr, "selftest: FAIL (%d failure(s))\n", failures)
	return 1
}

func main() {
	var skipVerify, useStrict bool

	lib.StrictTLSFlag(&useStrict)
	flag.BoolVar(&skipVerify, "k", false, "don't verify certificates")
	flag.Parse()

	tcfg, err := lib.BaselineTLSConfig(skipVerify, useStrict)
	die.If(err)

	args := flag.Args()

	if len(args) == 1 && args[0] == "selftest" {
		os.Exit(selftest())
	}

	if len(args) < 2 {
		fmt.Println("No certificates to check.")
		os.Exit(1)
	}

	caFile := args[0]
	args = args[1:]

	caCert, err := certlib.LoadCertificates(caFile)
	die.If(err)

	if len(caCert) != 1 {
		die.With("only one CA certificate should be presented.")
	}

	roots := x509.NewCertPool()
	roots.AddCert(caCert[0])

	for _, arg := range args {
		var cert *x509.Certificate

		cert, err = lib.GetCertificate(arg, tcfg)
		if err != nil {
			lib.Warn(err, "while parsing certificate from %s", arg)
			continue
		}

		if bytes.Equal(cert.AuthorityKeyId, caCert[0].AuthorityKeyId) {
			fmt.Printf("%s: SELF-SIGNED\n", arg)
			continue
		}

		if _, err = verify.CertWith(cert, roots, nil, false); err != nil {
			fmt.Printf("%s: INVALID\n", arg)
		} else {
			fmt.Printf("%s: OK (expires %s)\n", arg, cert.NotAfter.Format(lib.DateShortFormat))
		}
	}
}
