package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
	hosts "git.wntrmute.dev/kyle/goutils/certlib/hosts"
	"git.wntrmute.dev/kyle/goutils/certlib/revoke"
	"git.wntrmute.dev/kyle/goutils/fileutil"
)

var (
	hardfail bool
	timeout  time.Duration
	verbose  bool
)

var (
	strOK      = "OK"
	strExpired = "EXPIRED"
	strRevoked = "REVOKED"
	strUnknown = "UNKNOWN"
)

func main() {
	flag.BoolVar(&hardfail, "hardfail", false, "treat revocation check failures as fatal")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "network timeout for OCSP/CRL fetches and TLS site connects")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()

	revoke.HardFail = hardfail
	// Set HTTP client timeout for revocation library
	revoke.HTTPClient.Timeout = timeout

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <target> [<target>...]\n", os.Args[0])
		os.Exit(2)
	}

	exitCode := 0
	for _, target := range flag.Args() {
		status, err := processTarget(target)
		switch status {
		case strOK:
			fmt.Printf("%s: %s\n", target, strOK)
		case strExpired:
			fmt.Printf("%s: %s: %v\n", target, strExpired, err)
			exitCode = 1
		case strRevoked:
			fmt.Printf("%s: %s\n", target, strRevoked)
			exitCode = 1
		case strUnknown:
			fmt.Printf("%s: %s: %v\n", target, strUnknown, err)
			if hardfail {
				// In hardfail, treat unknown as failure
				exitCode = 1
			}
		}
	}

	os.Exit(exitCode)
}

func processTarget(target string) (string, error) {
	if fileutil.FileDoesExist(target) {
		return checkFile(target)
	}

	return checkSite(target)
}

func checkFile(path string) (string, error) {
	// Prefer high-level helpers from certlib to load certificates from disk
	if certs, err := certlib.LoadCertificates(path); err == nil && len(certs) > 0 {
		// Evaluate the first certificate (leaf) by default
		return evaluateCert(certs[0])
	}

	cert, err := certlib.LoadCertificate(path)
	if err != nil || cert == nil {
		return strUnknown, err
	}
	return evaluateCert(cert)
}

func checkSite(hostport string) (string, error) {
	// Use certlib/hosts to parse host/port (supports https URLs and host:port)
	target, err := hosts.ParseHost(hostport)
	if err != nil {
		return strUnknown, err
	}

	d := &net.Dialer{Timeout: timeout}
	tcfg := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         target.Host,
	} // #nosec G402 -- CLI tool only verifies revocation
	td := &tls.Dialer{NetDialer: d, Config: tcfg}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := td.DialContext(ctx, "tcp", target.String())
	if err != nil {
		return strUnknown, err
	}
	defer conn.Close()

	tconn, ok := conn.(*tls.Conn)
	if !ok {
		return strUnknown, errors.New("connection is not TLS")
	}

	state := tconn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return strUnknown, errors.New("no peer certificates presented")
	}
	return evaluateCert(state.PeerCertificates[0])
}

func evaluateCert(cert *x509.Certificate) (string, error) {
	// Delegate validity and revocation checks to certlib/revoke helper.
	// It returns revoked=true for both revoked and expired/not-yet-valid.
	// Map those cases back to our statuses using the returned error text.
	revoked, ok, err := revoke.VerifyCertificateError(cert)
	if revoked {
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "expired") || strings.Contains(msg, "isn't valid until") ||
				strings.Contains(msg, "not valid until") {
				return strExpired, err
			}
		}
		return strRevoked, err
	}
	if !ok {
		// Revocation status could not be determined
		return strUnknown, err
	}

	return strOK, nil
}
