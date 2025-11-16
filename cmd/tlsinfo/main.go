package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s ‹hostname:port>\n", os.Args[0])
		os.Exit(1)
	}

	hostPort := os.Args[1]
	conn, err := tls.Dial("tcp", hostPort, &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		fmt.Printf("Failed to connect to the TLS server: %v\n", err)
		os.Exit(1)
	}
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
