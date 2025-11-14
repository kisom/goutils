package main

import (
	"crypto/x509"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
)

// loadCertsFromFile attempts to parse certificates from a file that may be in
// PEM or DER/PKCS#7 format. Returns the parsed certificates or an error.
func loadCertsFromFile(path string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try PEM first
	if certs, err := certlib.ParseCertificatesPEM(data); err == nil {
		return certs, nil
	}

	// Try DER/PKCS7/PKCS12 (with no password)
	if certs, _, err := certlib.ParseCertificatesDER(data, ""); err == nil {
		return certs, nil
	} else {
		return nil, err
	}
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
	// Try PEM first
	if certs, err := certlib.ParseCertificatesPEM(data); err == nil {
		return certs, nil
	}
	// Try DER/PKCS7/PKCS12 (with no password)
	if certs, _, err := certlib.ParseCertificatesDER(data, ""); err == nil {
		return certs, nil
	} else {
		return nil, err
	}
}

func makePoolFromBytes(data []byte) (*x509.CertPool, error) {
	certs, err := loadCertsFromBytes(data)
	if err != nil || len(certs) == 0 {
		return nil, fmt.Errorf("failed to load CA certificates from embedded bytes")
	}
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool, nil
}

func verifyAgainstCA(caPool *x509.CertPool, path string) (ok bool, expiry string) {
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
	if _, err := leaf.Verify(opts); err != nil {
		return false, ""
	}

	return true, leaf.NotAfter.Format("2006-01-02")
}

func verifyAgainstCABytes(caPool *x509.CertPool, certData []byte) (ok bool, expiry string) {
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
	if _, err := leaf.Verify(opts); err != nil {
		return false, ""
	}

	return true, leaf.NotAfter.Format("2006-01-02")
}

// selftest runs built-in validation using embedded certificates.
func selftest() int {
	type testCase struct {
		name     string
		caFile   string
		certFile string
		expectOK bool
	}

	cases := []testCase{
		{name: "ISRG Root X1 validates LE E7", caFile: "testdata/isrg-root-x1.pem", certFile: "testdata/le-e7.pem", expectOK: true},
		{name: "ISRG Root X1 does NOT validate Google WR2", caFile: "testdata/isrg-root-x1.pem", certFile: "testdata/goog-wr2.pem", expectOK: false},
		{name: "GTS R1 validates Google WR2", caFile: "testdata/gts-r1.pem", certFile: "testdata/goog-wr2.pem", expectOK: true},
		{name: "GTS R1 does NOT validate LE E7", caFile: "testdata/gts-r1.pem", certFile: "testdata/le-e7.pem", expectOK: false},
	}

	failures := 0
	for _, tc := range cases {
		caBytes, err := embeddedTestdata.ReadFile(tc.caFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selftest: failed to read embedded %s: %v\n", tc.caFile, err)
			failures++
			continue
		}
		certBytes, err := embeddedTestdata.ReadFile(tc.certFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "selftest: failed to read embedded %s: %v\n", tc.certFile, err)
			failures++
			continue
		}
		pool, err := makePoolFromBytes(caBytes)
		if err != nil || pool == nil {
			fmt.Fprintf(os.Stderr, "selftest: failed to build CA pool for %s: %v\n", tc.caFile, err)
			failures++
			continue
		}
		ok, exp := verifyAgainstCABytes(pool, certBytes)
		if ok != tc.expectOK {
			fmt.Printf("%s: unexpected result: got %v, want %v\n", tc.name, ok, tc.expectOK)
			failures++
		} else {
			if ok {
				fmt.Printf("%s: OK (expires %s)\n", tc.name, exp)
			} else {
				fmt.Printf("%s: INVALID (as expected)\n", tc.name)
			}
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
		ok, exp := verifyAgainstCA(caPool, certPath)
		name := filepath.Base(certPath)
		if ok {
			// Display with the requested format
			// Example: file: OK (expires 2031-01-01)
			// Ensure deterministic date formatting
			// Note: no timezone displayed; date only as per example
			// If exp ended up empty for some reason, recompute safely
			if exp == "" {
				if certs, err := loadCertsFromFile(certPath); err == nil && len(certs) > 0 {
					exp = certs[0].NotAfter.Format("2006-01-02")
				} else {
					// fallback to the current date to avoid empty; though shouldn't happen
					exp = time.Now().Format("2006-01-02")
				}
			}
			fmt.Printf("%s: OK (expires %s)\n", name, exp)
		} else {
			fmt.Printf("%s: INVALID\n", name)
		}
	}
}
