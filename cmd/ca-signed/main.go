package main

import (
	"crypto/x509"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
)

// loadCertsFromFile attempts to parse certificates from a file that may be in
// PEM or DER/PKCS#7 format. Returns the parsed certificates or an error.
func loadCertsFromFile(path string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if certs, err = certlib.ParseCertificatesPEM(data); err == nil {
		return certs, nil
	}

	if certs, _, err = certlib.ParseCertificatesDER(data, ""); err == nil {
		return certs, nil
	}

	return nil, err
}

func makePoolFromFile(path string) (*x509.CertPool, error) {
	// Try PEM via helper (it builds a pool)
	if pool, err := certlib.LoadPEMCertPool(path); err == nil && pool != nil {
		return pool, nil
	}

	// Fallback: read as DER(s), add to a new pool
	certs, err := loadCertsFromFile(path)
	if err != nil || len(certs) == 0 {
		return nil, fmt.Errorf("failed to load CA certificates from %s", path)
	}
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool, nil
}

//go:embed testdata/*.pem
var embeddedTestdata embed.FS

// loadCertsFromBytes attempts to parse certificates from bytes that may be in
// PEM or DER/PKCS#7 format.
func loadCertsFromBytes(data []byte) ([]*x509.Certificate, error) {
	certs, err := certlib.ParseCertificatesPEM(data)
	if err == nil {
		return certs, nil
	}

	certs, _, err = certlib.ParseCertificatesDER(data, "")
	if err == nil {
		return certs, nil
	}

	return nil, err
}

func makePoolFromBytes(data []byte) (*x509.CertPool, error) {
	certs, err := loadCertsFromBytes(data)
	if err != nil || len(certs) == 0 {
		return nil, errors.New("failed to load CA certificates from embedded bytes")
	}
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool, nil
}

// isSelfSigned returns true if the given certificate is self-signed.
// It checks that the subject and issuer match and that the certificate's
// signature verifies against its own public key.
func isSelfSigned(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	// Quick check: subject and issuer match
	if cert.Subject.String() != cert.Issuer.String() {
		return false
	}
	// Cryptographic check: the certificate is signed by itself
	if err := cert.CheckSignatureFrom(cert); err != nil {
		return false
	}
	return true
}

func verifyAgainstCA(caPool *x509.CertPool, path string) (bool, string) {
	certs, err := loadCertsFromFile(path)
	if err != nil || len(certs) == 0 {
		return false, ""
	}

	leaf := certs[0]
	ints := x509.NewCertPool()
	if len(certs) > 1 {
		for _, ic := range certs[1:] {
			ints.AddCert(ic)
		}
	}

	opts := x509.VerifyOptions{
		Roots:         caPool,
		Intermediates: ints,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	if _, err = leaf.Verify(opts); err != nil {
		return false, ""
	}

	return true, leaf.NotAfter.Format("2006-01-02")
}

func verifyAgainstCABytes(caPool *x509.CertPool, certData []byte) (bool, string) {
	certs, err := loadCertsFromBytes(certData)
	if err != nil || len(certs) == 0 {
		return false, ""
	}

	leaf := certs[0]
	ints := x509.NewCertPool()
	if len(certs) > 1 {
		for _, ic := range certs[1:] {
			ints.AddCert(ic)
		}
	}

	opts := x509.VerifyOptions{
		Roots:         caPool,
		Intermediates: ints,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	if _, err = leaf.Verify(opts); err != nil {
		return false, ""
	}

	return true, leaf.NotAfter.Format("2006-01-02")
}

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

	pool, err := makePoolFromBytes(caBytes)
	if err != nil || pool == nil {
		return fmt.Errorf("selftest: failed to build CA pool for %s: %w", tc.caFile, err)
	}

	ok, exp := verifyAgainstCABytes(pool, certBytes)
	if ok != tc.expectOK {
		return fmt.Errorf("%s: unexpected result: got %v, want %v", tc.name, ok, tc.expectOK)
	}

	if ok {
		fmt.Printf("%s: OK (expires %s)\n", tc.name, exp)
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
		certs, err := loadCertsFromBytes(b)
		if err != nil || len(certs) == 0 {
			fmt.Fprintf(os.Stderr, "selftest: failed to parse cert(s) from %s: %v\n", root, err)
			failures++
			continue
		}
		leaf := certs[0]
		if isSelfSigned(leaf) {
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

// expiryString returns a YYYY-MM-DD date string to display for certificate
// expiry. If an explicit exp string is provided, it is used. Otherwise, if a
// leaf certificate is available, its NotAfter is formatted. As a last resort,
// it falls back to today's date (should not normally happen).
func expiryString(leaf *x509.Certificate, exp string) string {
	if exp != "" {
		return exp
	}
	if leaf != nil {
		return leaf.NotAfter.Format("2006-01-02")
	}
	return time.Now().Format("2006-01-02")
}

// processCert verifies a single certificate file against the provided CA pool
// and prints the result in the required format, handling self-signed
// certificates specially.
func processCert(caPool *x509.CertPool, certPath string) {
	ok, exp := verifyAgainstCA(caPool, certPath)
	name := filepath.Base(certPath)

	// Try to load the leaf cert for self-signed detection and expiry fallback
	var leaf *x509.Certificate
	if certs, err := loadCertsFromFile(certPath); err == nil && len(certs) > 0 {
		leaf = certs[0]
	}

	// Prefer the SELF-SIGNED label if applicable
	if isSelfSigned(leaf) {
		fmt.Printf("%s: SELF-SIGNED\n", name)
		return
	}

	if ok {
		fmt.Printf("%s: OK (expires %s)\n", name, expiryString(leaf, exp))
		return
	}
	fmt.Printf("%s: INVALID\n", name)
}

func main() {
	// Special selftest mode: single argument "selftest"
	if len(os.Args) == 2 && os.Args[1] == "selftest" {
		os.Exit(selftest())
	}

	if len(os.Args) < 3 {
		prog := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage:\n  %s ca.pem cert1.pem cert2.pem ...\n  %s selftest\n", prog, prog)
		os.Exit(2)
	}

	caPath := os.Args[1]
	caPool, err := makePoolFromFile(caPath)
	if err != nil || caPool == nil {
		fmt.Fprintf(os.Stderr, "failed to load CA certificate(s): %v\n", err)
		os.Exit(1)
	}

	for _, certPath := range os.Args[2:] {
		processCert(caPool, certPath)
	}
}
