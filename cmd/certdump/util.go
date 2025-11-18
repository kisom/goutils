package main

import (
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/kr/text"
)

// following two lifted from CFSSL, (replace-regexp "\(.+\): \(.+\),"
// "\2: \1,")

const (
	sSHA256 = "SHA256"
	sSHA512 = "SHA512"
)

var keyUsage = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "digital signature",
	x509.KeyUsageContentCommitment: "content committment",
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
	var s string
	var sSb130 strings.Builder
	for i := range in {
		sSb130.WriteString(fmt.Sprintf("%02X:", in[i]))
	}
	s += sSb130.String()

	return strings.Trim(s, ":")
}
