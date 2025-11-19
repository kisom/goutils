package main

import (
	"crypto/x509"
	"flag"
	"fmt"

	"git.wntrmute.dev/kyle/goutils/certlib/verify"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
	"git.wntrmute.dev/kyle/goutils/lib/dialer"
	"git.wntrmute.dev/kyle/goutils/lib/fetch"
)

func main() {
	var (
		skipVerify bool
		strictTLS  bool
		leeway     = verify.DefaultLeeway
		warnOnly   bool
	)

	dialer.StrictTLSFlag(&strictTLS)

	flag.BoolVar(&skipVerify, "k", false, "skip server verification") // #nosec G402
	flag.BoolVar(&warnOnly, "q", false, "only warn about expiring certs")
	flag.DurationVar(&leeway, "t", leeway, "warn if certificates are closer than this to expiring")
	flag.Parse()

	tlsCfg, err := dialer.BaselineTLSConfig(skipVerify, strictTLS)
	die.If(err)

	for _, file := range flag.Args() {
		var certs []*x509.Certificate

		certs, err = fetch.GetCertificateChain(file, tlsCfg)
		if err != nil {
			_, _ = lib.Warn(err, "while parsing certificates")
			continue
		}

		for _, cert := range certs {
			check := verify.NewCertCheck(cert, leeway)

			if warnOnly {
				if err = check.Err(); err != nil {
					lib.Warn(err, "certificate is expiring")
				}
			} else {
				fmt.Printf("%s expires on %s (in %s)\n", check.Name(),
					cert.NotAfter, check.Expiry())
			}
		}
	}
}
