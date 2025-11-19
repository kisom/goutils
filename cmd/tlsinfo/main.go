package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib/hosts"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s ‹hostname:port>\n", os.Args[0])
		os.Exit(1)
	}

	hostPort, err := hosts.ParseHost(os.Args[1])
	die.If(err)

	// Use proxy-aware TLS dialer; skip verification as before
	conn, err := lib.DialTLS(
		context.Background(),
		hostPort.String(),
		lib.DialerOpts{TLSConfig: &tls.Config{InsecureSkipVerify: true}},
	) // #nosec G402
	die.If(err)

	defer conn.Close()

	state := conn.ConnectionState()
	printConnectionDetails(state)
}

func printConnectionDetails(state tls.ConnectionState) {
	version := tlsVersion(state.Version)
	cipherSuite := tls.CipherSuiteName(state.CipherSuite)
	fmt.Printf("TLS Version: %s\n", version)
	fmt.Printf("Cipher Suite: %s\n", cipherSuite)
	printPeerCertificates(state.PeerCertificates)
}

func tlsVersion(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:

		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return "Unknown"
	}
}

func printPeerCertificates(certificates []*x509.Certificate) {
	for i, cert := range certificates {
		fmt.Printf("Certificate %d\n", i+1)
		fmt.Printf("\tSubject: %s\n", cert.Subject)
		fmt.Printf("\tIssuer: %s\n", cert.Issuer)
		fmt.Printf("\tDNS Names:	%v\n", cert.DNSNames)
		fmt.Printf("\tNot Before: %s\n:", cert.NotBefore)
		fmt.Printf("\tNot After: %s\n", cert.NotAfter)
	}
}
