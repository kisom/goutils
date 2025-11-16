package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/die"
	"git.wntrmute.dev/kyle/goutils/lib"
)

var warnOnly bool
var leeway = 2160 * time.Hour // three months

func displayName(name pkix.Name) string {
	var ns []string

	if name.CommonName != "" {
		ns = append(ns, name.CommonName)
	}

	for i := range name.Country {
		ns = append(ns, fmt.Sprintf("C=%s", name.Country[i]))
	}

	for i := range name.Organization {
		ns = append(ns, fmt.Sprintf("O=%s", name.Organization[i]))
	}

	for i := range name.OrganizationalUnit {
		ns = append(ns, fmt.Sprintf("OU=%s", name.OrganizationalUnit[i]))
	}

	for i := range name.Locality {
		ns = append(ns, fmt.Sprintf("L=%s", name.Locality[i]))
	}

	for i := range name.Province {
		ns = append(ns, fmt.Sprintf("ST=%s", name.Province[i]))
	}

	if len(ns) > 0 {
		return "/" + strings.Join(ns, "/")
	}

	die.With("no subject information in root")
	return ""
}

func expires(cert *x509.Certificate) time.Duration {
	return time.Until(cert.NotAfter)
}

func inDanger(cert *x509.Certificate) bool {
	return expires(cert) < leeway
}

func checkCert(cert *x509.Certificate) {
	warn := inDanger(cert)
	name := displayName(cert.Subject)
	name = fmt.Sprintf("%s/SN=%s", name, cert.SerialNumber)
	expiry := expires(cert)
	if warnOnly {
		if warn {
			fmt.Fprintf(os.Stderr, "%s expires on %s (in %s)\n", name, cert.NotAfter, expiry)
		}
	} else {
		fmt.Printf("%s expires on %s (in %s)\n", name, cert.NotAfter, expiry)
	}
}

func main() {
	flag.BoolVar(&warnOnly, "q", false, "only warn about expiring certs")
	flag.DurationVar(&leeway, "t", leeway, "warn if certificates are closer than this to expiring")
	flag.Parse()

	for _, file := range flag.Args() {
		in, err := os.ReadFile(file)
		if err != nil {
			_, _ = lib.Warn(err, "failed to read file")
			continue
		}

		certs, err := certlib.ParseCertificatesPEM(in)
		if err != nil {
			_, _ = lib.Warn(err, "while parsing certificates")
			continue
		}

		for _, cert := range certs {
			checkCert(cert)
		}
	}
}
