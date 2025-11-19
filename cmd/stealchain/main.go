package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib/dialer"
)

func main() {
	var sysRoot, serverName string
	var skipVerify bool
	var strictTLS bool
	dialer.StrictTLSFlag(&strictTLS)
	flag.StringVar(&sysRoot, "ca", "", "provide an alternate CA bundle")
	flag.StringVar(&serverName, "sni", "", "provide an SNI name")
	flag.BoolVar(&skipVerify, "noverify", false, "don't verify certificates")
	flag.Parse()

	tlsCfg, err := dialer.BaselineTLSConfig(skipVerify, strictTLS)
	die.If(err)

	if sysRoot != "" {
		tlsCfg.RootCAs, err = certlib.LoadPEMCertPool(sysRoot)
		die.If(err)
	}

	if serverName != "" {
		tlsCfg.ServerName = serverName
	}

	for _, site := range flag.Args() {
		_, _, err = net.SplitHostPort(site)
		if err != nil {
			site += ":443"
		}

		var conn *tls.Conn
		conn, err = dialer.DialTLS(context.Background(), site, dialer.Opts{TLSConfig: tlsCfg})
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
