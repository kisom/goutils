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
	"git.wntrmute.dev/kyle/goutils/lib/dialer"
)

var hasPort = regexp.MustCompile(`:\d+$`)

func main() {
	flag.Parse()

	for _, server := range flag.Args() {
		if !hasPort.MatchString(server) {
			server += ":443"
		}

		// Use proxy-aware TLS dialer
		conn, err := dialer.DialTLS(
			context.Background(),
			server,
			dialer.Opts{TLSConfig: &tls.Config{}},
		) // #nosec G402
		die.If(err)

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
