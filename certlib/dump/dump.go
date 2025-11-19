// Package dump implements tooling for dumping certificate information.
package dump

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/kr/text"

	"git.wntrmute.dev/kyle/goutils/lib"
)

const (
	sSHA256 = "SHA256"
	sSHA512 = "SHA512"
)

var keyUsage = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "digital signature",
	x509.KeyUsageContentCommitment: "content commitment",
	x509.KeyUsageKeyEncipherment:   "key encipherment",
	x509.KeyUsageKeyAgreement:      "key agreement",
	x509.KeyUsageDataEncipherment:  "data encipherment",
	x509.KeyUsageCertSign:          "cert sign",
	x509.KeyUsageCRLSign:           "crl sign",
	x509.KeyUsageEncipherOnly:      "encipher only",
	x509.KeyUsageDecipherOnly:      "decipher only",
}

var extKeyUsages = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                            "any",
	x509.ExtKeyUsageServerAuth:                     "server auth",
	x509.ExtKeyUsageClientAuth:                     "client auth",
	x509.ExtKeyUsageCodeSigning:                    "code signing",
	x509.ExtKeyUsageEmailProtection:                "s/mime",
	x509.ExtKeyUsageIPSECEndSystem:                 "ipsec end system",
	x509.ExtKeyUsageIPSECTunnel:                    "ipsec tunnel",
	x509.ExtKeyUsageIPSECUser:                      "ipsec user",
	x509.ExtKeyUsageTimeStamping:                   "timestamping",
	x509.ExtKeyUsageOCSPSigning:                    "ocsp signing",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "microsoft sgc",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:      "netscape sgc",
	x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "microsoft commercial code signing",
	x509.ExtKeyUsageMicrosoftKernelCodeSigning:     "microsoft kernel code signing",
}

func sigAlgoPK(a x509.SignatureAlgorithm) string {
	switch a {
	case x509.MD2WithRSA, x509.MD5WithRSA, x509.SHA1WithRSA, x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA:
		return "RSA"
	case x509.SHA256WithRSAPSS, x509.SHA384WithRSAPSS, x509.SHA512WithRSAPSS:
		return "RSA-PSS"
	case x509.ECDSAWithSHA1, x509.ECDSAWithSHA256, x509.ECDSAWithSHA384, x509.ECDSAWithSHA512:
		return "ECDSA"
	case x509.DSAWithSHA1, x509.DSAWithSHA256:
		return "DSA"
	case x509.PureEd25519:
		return "Ed25519"
	case x509.UnknownSignatureAlgorithm:
		return "unknown public key algorithm"
	default:
		return "unknown public key algorithm"
	}
}

func sigAlgoHash(a x509.SignatureAlgorithm) string {
	switch a {
	case x509.MD2WithRSA:
		return "MD2"
	case x509.MD5WithRSA:
		return "MD5"
	case x509.SHA1WithRSA, x509.ECDSAWithSHA1, x509.DSAWithSHA1:
		return "SHA1"
	case x509.SHA256WithRSA, x509.ECDSAWithSHA256, x509.DSAWithSHA256:
		return sSHA256
	case x509.SHA256WithRSAPSS:
		return sSHA256
	case x509.SHA384WithRSA, x509.ECDSAWithSHA384:
		return "SHA384"
	case x509.SHA384WithRSAPSS:
		return "SHA384"
	case x509.SHA512WithRSA, x509.ECDSAWithSHA512:
		return sSHA512
	case x509.SHA512WithRSAPSS:
		return sSHA512
	case x509.PureEd25519:
		return sSHA512
	case x509.UnknownSignatureAlgorithm:
		return "unknown hash algorithm"
	default:
		return "unknown hash algorithm"
	}
}

const maxLine = 78

func makeIndent(n int) string {
	s := "    "
	var sSb97 strings.Builder
	for range n {
		sSb97.WriteString("        ")
	}
	s += sSb97.String()
	return s
}

func indentLen(n int) int {
	return 4 + (8 * n)
}

// this isn't real efficient, but that's not a problem here.
func wrap(s string, indent int) string {
	if indent > 3 {
		indent = 3
	}

	wrapped := text.Wrap(s, maxLine)
	lines := strings.SplitN(wrapped, "\n", 2)
	if len(lines) == 1 {
		return lines[0]
	}

	if (maxLine - indentLen(indent)) <= 0 {
		panic("too much indentation")
	}

	rest := strings.Join(lines[1:], " ")
	wrapped = text.Wrap(rest, maxLine-indentLen(indent))
	return lines[0] + "\n" + text.Indent(wrapped, makeIndent(indent))
}

func dumpHex(in []byte) string {
	return lib.HexEncode(in, lib.HexEncodeUpperColon)
}

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

func DisplayName(name pkix.Name) string {
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
		fmt.Fprint(os.Stdout, ", is not a CA certificate")
		if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
			fmt.Fprint(os.Stdout, " (key encipherment usage enabled!)")
		}
	}

	if (cert.MaxPathLen == 0 && cert.MaxPathLenZero) || (cert.MaxPathLen > 0) {
		fmt.Fprintf(os.Stdout, ", max path length %d", cert.MaxPathLen)
	}

	fmt.Fprintln(os.Stdout)
}

func wrapPrint(text string, indent int) {
	tabs := ""
	var tabsSb140 strings.Builder
	for range indent {
		tabsSb140.WriteString("\t")
	}
	tabs += tabsSb140.String()

	fmt.Fprintf(os.Stdout, tabs+"%s\n", wrap(text, indent))
}

func DisplayCert(w io.Writer, cert *x509.Certificate, showHash bool) {
	fmt.Fprintln(w, "CERTIFICATE")
	if showHash {
		fmt.Fprintln(w, wrap(fmt.Sprintf("SHA256: %x", sha256.Sum256(cert.Raw)), 0))
	}

	fmt.Fprintln(w, wrap("Subject: "+DisplayName(cert.Subject), 0))
	fmt.Fprintln(w, wrap("Issuer: "+DisplayName(cert.Issuer), 0))
	fmt.Fprintf(w, "\tSignature algorithm: %s / %s\n", sigAlgoPK(cert.SignatureAlgorithm),
		sigAlgoHash(cert.SignatureAlgorithm))
	fmt.Fprintln(w, "Details:")
	wrapPrint("Public key: "+certPublic(cert), 1)
	fmt.Fprintf(w, "\tSerial number: %s\n", cert.SerialNumber)

	if len(cert.AuthorityKeyId) > 0 {
		fmt.Fprintf(w, "\t%s\n", wrap("AKI: "+dumpHex(cert.AuthorityKeyId), 1))
	}
	if len(cert.SubjectKeyId) > 0 {
		fmt.Fprintf(w, "\t%s\n", wrap("SKI: "+dumpHex(cert.SubjectKeyId), 1))
	}

	wrapPrint("Valid from: "+cert.NotBefore.Format(lib.DateShortFormat), 1)
	fmt.Fprintf(w, "\t     until: %s\n", cert.NotAfter.Format(lib.DateShortFormat))
	fmt.Fprintf(w, "\tKey usages: %s\n", keyUsages(cert.KeyUsage))

	if len(cert.ExtKeyUsage) > 0 {
		fmt.Fprintf(w, "\tExtended usages: %s\n", extUsage(cert.ExtKeyUsage))
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
