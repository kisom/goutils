package main

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"git.wntrmute.dev/kyle/goutils/die"
)

var hasPort = regexp.MustCompile(`:\d+$`)

func main() {
	flag.Parse()

	for _, server := range flag.Args() {
		if !hasPort.MatchString(server) {
			server += ":443"
		}

		d := &tls.Dialer{Config: &tls.Config{}} // #nosec G402
		nc, err := d.DialContext(context.Background(), "tcp", server)
		die.If(err)
		conn, ok := nc.(*tls.Conn)
		if !ok {
			die.With("invalid TLS connection (not a *tls.Conn)")
		}

		defer conn.Close()

		details := conn.ConnectionState()
		var chain strings.Builder
		for _, cert := range details.PeerCertificates {
			p := pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			}
			chain.Write(pem.EncodeToMemory(&p))
		}

		fmt.Fprintln(os.Stdout, chain.String())
	}
}
