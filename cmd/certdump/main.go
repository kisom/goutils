//lint:file-ignore SA1019 allow strict compatibility for old certs
package main

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"git.wntrmute.dev/kyle/goutils/certlib"
	"git.wntrmute.dev/kyle/goutils/lib"
)

func certPublic(cert *x509.Certificate) string {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA-%d", pub.N.BitLen())
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P256():
			return "ECDSA-prime256v1"
		case elliptic.P384():
			return "ECDSA-secp384r1"
		case elliptic.P521():
			return "ECDSA-secp521r1"
		default:
			return "ECDSA (unknown curve)"
		}
	case *dsa.PublicKey:
		return "DSA"
	default:
		return "Unknown"
	}
}

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

	return "*** no subject information ***"
}

func keyUsages(ku x509.KeyUsage) string {
	var uses []string

	for u, s := range keyUsage {
		if (ku & u) != 0 {
			uses = append(uses, s)
		}
	}
	sort.Strings(uses)

	return strings.Join(uses, ", ")
}

func extUsage(ext []x509.ExtKeyUsage) string {
	ns := make([]string, 0, len(ext))
	for i := range ext {
		ns = append(ns, extKeyUsages[ext[i]])
	}
	sort.Strings(ns)

	return strings.Join(ns, ", ")
}

func showBasicConstraints(cert *x509.Certificate) {
	fmt.Fprint(os.Stdout, "\tBasic constraints: ")
	if cert.BasicConstraintsValid {
		fmt.Fprint(os.Stdout, "valid")
	} else {
		fmt.Fprint(os.Stdout, "invalid")
	}

	if cert.IsCA {
		fmt.Fprint(os.Stdout, ", is a CA certificate")
		if !cert.BasicConstraintsValid {
			fmt.Fprint(os.Stdout, " (basic constraint failure)")
		}
	} else {
		fmt.Fprint(os.Stdout, "is not a CA certificate")
		if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
			fmt.Fprint(os.Stdout, " (key encipherment usage enabled!)")
		}
	}

	if (cert.MaxPathLen == 0 && cert.MaxPathLenZero) || (cert.MaxPathLen > 0) {
		fmt.Fprintf(os.Stdout, ", max path length %d", cert.MaxPathLen)
	}

	fmt.Fprintln(os.Stdout)
}

const oneTrueDateFormat = "2006-01-02T15:04:05-0700"

var (
	dateFormat string
	showHash   bool // if true, print a SHA256 hash of the certificate's Raw field
)

func wrapPrint(text string, indent int) {
	tabs := ""
	var tabsSb140 strings.Builder
	for range indent {
		tabsSb140.WriteString("\t")
	}
	tabs += tabsSb140.String()

	fmt.Fprintf(os.Stdout, tabs+"%s\n", wrap(text, indent))
}

func displayCert(cert *x509.Certificate) {
	fmt.Fprintln(os.Stdout, "CERTIFICATE")
	if showHash {
		fmt.Fprintln(os.Stdout, wrap(fmt.Sprintf("SHA256: %x", sha256.Sum256(cert.Raw)), 0))
	}
	fmt.Fprintln(os.Stdout, wrap("Subject: "+displayName(cert.Subject), 0))
	fmt.Fprintln(os.Stdout, wrap("Issuer: "+displayName(cert.Issuer), 0))
	fmt.Fprintf(os.Stdout, "\tSignature algorithm: %s / %s\n", sigAlgoPK(cert.SignatureAlgorithm),
		sigAlgoHash(cert.SignatureAlgorithm))
	fmt.Fprintln(os.Stdout, "Details:")
	wrapPrint("Public key: "+certPublic(cert), 1)
	fmt.Fprintf(os.Stdout, "\tSerial number: %s\n", cert.SerialNumber)

	if len(cert.AuthorityKeyId) > 0 {
		fmt.Fprintf(os.Stdout, "\t%s\n", wrap("AKI: "+dumpHex(cert.AuthorityKeyId), 1))
	}
	if len(cert.SubjectKeyId) > 0 {
		fmt.Fprintf(os.Stdout, "\t%s\n", wrap("SKI: "+dumpHex(cert.SubjectKeyId), 1))
	}

	wrapPrint("Valid from: "+cert.NotBefore.Format(dateFormat), 1)
	fmt.Fprintf(os.Stdout, "\t     until: %s\n", cert.NotAfter.Format(dateFormat))
	fmt.Fprintf(os.Stdout, "\tKey usages: %s\n", keyUsages(cert.KeyUsage))

	if len(cert.ExtKeyUsage) > 0 {
		fmt.Fprintf(os.Stdout, "\tExtended usages: %s\n", extUsage(cert.ExtKeyUsage))
	}

	showBasicConstraints(cert)

	validNames := make([]string, 0, len(cert.DNSNames)+len(cert.EmailAddresses)+len(cert.IPAddresses))
	for i := range cert.DNSNames {
		validNames = append(validNames, "dns:"+cert.DNSNames[i])
	}

	for i := range cert.EmailAddresses {
		validNames = append(validNames, "email:"+cert.EmailAddresses[i])
	}

	for i := range cert.IPAddresses {
		validNames = append(validNames, "ip:"+cert.IPAddresses[i].String())
	}

	sans := fmt.Sprintf("SANs (%d): %s\n", len(validNames), strings.Join(validNames, ", "))
	wrapPrint(sans, 1)

	l := len(cert.IssuingCertificateURL)
	if l != 0 {
		var aia string
		if l == 1 {
			aia = "AIA"
		} else {
			aia = "AIAs"
		}
		wrapPrint(fmt.Sprintf("%d %s:", l, aia), 1)
		for _, url := range cert.IssuingCertificateURL {
			wrapPrint(url, 2)
		}
	}

	l = len(cert.OCSPServer)
	if l > 0 {
		title := "OCSP server"
		if l > 1 {
			title += "s"
		}
		wrapPrint(title+":\n", 1)
		for _, ocspServer := range cert.OCSPServer {
			wrapPrint(fmt.Sprintf("- %s\n", ocspServer), 2)
		}
	}
}

func main() {
	var leafOnly bool
	flag.BoolVar(&showHash, "d", false, "show hashes of raw DER contents")
	flag.StringVar(&dateFormat, "s", oneTrueDateFormat, "date `format` in Go time format")
	flag.BoolVar(&leafOnly, "l", false, "only show the leaf certificate")
	flag.Parse()

	opts := &certlib.FetcherOpts{
		SkipVerify: true,
		Roots:      nil,
	}

	for _, filename := range flag.Args() {
		fmt.Fprintf(os.Stdout, "--%s ---%s", filename, "\n")
		certs, err := certlib.GetCertificateChain(filename, opts)
		if err != nil {
			_, _ = lib.Warn(err, "couldn't read certificate")
			continue
		}

		if leafOnly {
			displayCert(certs[0])
			continue
		}

		for i := range certs {
			displayCert(certs[i])
		}
	}
}
