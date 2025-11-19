package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"os"

	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func main() {
	var cfg = &tls.Config{} // #nosec G402

	var sysRoot, serverName string
	flag.StringVar(&sysRoot, "ca", "", "provide an alternate CA bundle")
	flag.StringVar(&cfg.ServerName, "sni", cfg.ServerName, "provide an SNI name")
	flag.BoolVar(&cfg.InsecureSkipVerify, "noverify", false, "don't verify certificates")
	flag.Parse()

	if sysRoot != "" {
		pemList, err := os.ReadFile(sysRoot)
		die.If(err)

		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(pemList) {
			fmt.Printf("[!] no valid roots found")
			roots = nil
		}

		cfg.RootCAs = roots
	}

	if serverName != "" {
		cfg.ServerName = serverName
	}

	for _, site := range flag.Args() {
		_, _, err := net.SplitHostPort(site)
		if err != nil {
			site += ":443"
		}
		// Use proxy-aware TLS dialer
		conn, err := lib.DialTLS(context.Background(), site, lib.DialerOpts{TLSConfig: cfg})
		die.If(err)

		cs := conn.ConnectionState()
		var chain []byte

		for _, cert := range cs.PeerCertificates {
			p := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			}
			chain = append(chain, pem.EncodeToMemory(p)...)
		}

		err = os.WriteFile(site+".pem", chain, 0644)
		die.If(err)

		fmt.Printf("[+] wrote %s.pem.\n", site)
	}
}
