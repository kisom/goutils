//lint:file-ignore SA1019 allow strict compatibility for old certs
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	"git.wntrmute.dev/kyle/goutils/certlib/dump"
	"git.wntrmute.dev/kyle/goutils/lib"
	"git.wntrmute.dev/kyle/goutils/lib/fetch"
)

var config struct {
	showHash   bool
	dateFormat string
	leafOnly   bool
}

func main() {
	flag.BoolVar(&config.showHash, "d", false, "show hashes of raw DER contents")
	flag.StringVar(&config.dateFormat, "s", lib.OneTrueDateFormat, "date `format` in Go time format")
	flag.BoolVar(&config.leafOnly, "l", false, "only show the leaf certificate")
	flag.Parse()

	tlsCfg := &tls.Config{InsecureSkipVerify: true} // #nosec G402 - tool intentionally inspects broken TLS

	for _, filename := range flag.Args() {
		fmt.Fprintf(os.Stdout, "--%s ---%s", filename, "\n")
		certs, err := fetch.GetCertificateChain(filename, tlsCfg)
		if err != nil {
			lib.Warn(err, "couldn't read certificate")
			continue
		}

		if config.leafOnly {
			dump.DisplayCert(os.Stdout, certs[0], config.showHash)
			continue
		}

		for i := range certs {
			dump.DisplayCert(os.Stdout, certs[i], config.showHash)
		}
	}
}
