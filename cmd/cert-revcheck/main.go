package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
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
		case "OK":
			fmt.Printf("%s: OK\n", target)
		case "EXPIRED":
			fmt.Printf("%s: EXPIRED: %v\n", target, err)
			exitCode = 1
		case "REVOKED":
			fmt.Printf("%s: REVOKED\n", target)
			exitCode = 1
		case "UNKNOWN":
			fmt.Printf("%s: UNKNOWN: %v\n", target, err)
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

	// Not a file; treat as site
	return checkSite(target)
}

func checkFile(path string) (string, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return "UNKNOWN", err
	}

	// Try PEM first; if that fails, try single DER cert
	certs, err := certlib.ReadCertificates(in)
	if err != nil || len(certs) == 0 {
		cert, _, derr := certlib.ReadCertificate(in)
		if derr != nil || cert == nil {
			if err == nil {
				err = derr
			}
			return "UNKNOWN", err
		}
		return evaluateCert(cert)
	}

	// Evaluate the first certificate (leaf) by default
	return evaluateCert(certs[0])
}

func checkSite(hostport string) (string, error) {
	// Use certlib/hosts to parse host/port (supports https URLs and host:port)
	target, err := hosts.ParseHost(hostport)
	if err != nil {
		return "UNKNOWN", err
	}

	d := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(d, "tcp", target.String(), &tls.Config{InsecureSkipVerify: true, ServerName: target.Host})
	if err != nil {
		return "UNKNOWN", err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return "UNKNOWN", errors.New("no peer certificates presented")
	}
	return evaluateCert(state.PeerCertificates[0])
}

func evaluateCert(cert *x509.Certificate) (string, error) {
	// Expiry check
	now := time.Now()
	if !now.Before(cert.NotAfter) {
		return "EXPIRED", fmt.Errorf("expired at %s", cert.NotAfter)
	}
	if !now.After(cert.NotBefore) {
		return "EXPIRED", fmt.Errorf("not valid until %s", cert.NotBefore)
	}

	// Revocation check using certlib/revoke
	revoked, ok, err := revoke.VerifyCertificateError(cert)
	if revoked {
		// If revoked is true, ok will be true per implementation, err may describe why
		return "REVOKED", err
	}
	if !ok {
		// Revocation status could not be determined
		return "UNKNOWN", err
	}

	return "OK", nil
}
