package certlib

// Originally from CFSSL, mostly written by me originally, and licensed under:

/*
Copyright (c) 2014 CloudFlare Inc.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions
are met:

Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// I've modified it for use in my own code e.g. by removing the CFSSL errors
// and replacing them with sane ones.

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	ct "github.com/google/certificate-transparency-go"
	cttls "github.com/google/certificate-transparency-go/tls"
	ctx509 "github.com/google/certificate-transparency-go/x509"
	"golang.org/x/crypto/ocsp"
	"golang.org/x/crypto/pkcs12"

	"git.wntrmute.dev/kyle/goutils/certlib/certerr"
	"git.wntrmute.dev/kyle/goutils/certlib/pkcs7"
)

// OneYear is a time.Duration representing a year's worth of seconds.
const OneYear = 8760 * time.Hour

// OneDay is a time.Duration representing a day's worth of seconds.
const OneDay = 24 * time.Hour

// DelegationUsage  is the OID for the DelegationUseage extensions.
var DelegationUsage = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 44363, 44}

// DelegationExtension is a non-critical extension marking delegation usage.
var DelegationExtension = pkix.Extension{
	Id:       DelegationUsage,
	Critical: false,
	Value:    []byte{0x05, 0x00}, // ASN.1 NULL
}

// InclusiveDate returns the time.Time representation of a date - 1
// nanosecond. This allows time.After to be used inclusively.
func InclusiveDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Add(-1 * time.Nanosecond)
}

const (
	year2012 = 2012
	year2015 = 2015
	day1     = 1
)

// Jul2012 is the July 2012 CAB Forum deadline for when CAs must stop
// issuing certificates valid for more than 5 years.
var Jul2012 = InclusiveDate(year2012, time.July, day1)

// Apr2015 is the April 2015 CAB Forum deadline for when CAs must stop
// issuing certificates valid for more than 39 months.
var Apr2015 = InclusiveDate(year2015, time.April, day1)

// KeyLength returns the bit size of ECDSA or RSA PublicKey.
func KeyLength(key any) int {
	switch k := key.(type) {
	case *ecdsa.PublicKey:
		if k == nil {
			return 0
		}
		return k.Curve.Params().BitSize
	case *rsa.PublicKey:
		if k == nil {
			return 0
		}
		return k.N.BitLen()
	default:
		return 0
	}
}

// ExpiryTime returns the time when the certificate chain is expired.
func ExpiryTime(chain []*x509.Certificate) time.Time {
	var notAfter time.Time
	if len(chain) == 0 {
		return notAfter
	}
	notAfter = chain[0].NotAfter
	for _, cert := range chain {
		if notAfter.After(cert.NotAfter) {
			notAfter = cert.NotAfter
		}
	}
	return notAfter
}

// MonthsValid returns the number of months for which a certificate is valid.
func MonthsValid(c *x509.Certificate) int {
	issued := c.NotBefore
	expiry := c.NotAfter
	years := (expiry.Year() - issued.Year())
	months := years*12 + int(expiry.Month()) - int(issued.Month())

	// Round up if valid for less than a full month
	if expiry.Day() > issued.Day() {
		months++
	}
	return months
}

// ValidExpiry determines if a certificate is valid for an acceptable
// length of time per the CA/Browser Forum baseline requirements.
// See https://cabforum.org/wp-content/uploads/CAB-Forum-BR-1.3.0.pdf
func ValidExpiry(c *x509.Certificate) bool {
	issued := c.NotBefore

	var maxMonths int
	switch {
	case issued.After(Apr2015):
		maxMonths = 39
	case issued.After(Jul2012):
		maxMonths = 60
	default:
		maxMonths = 120
	}

	return MonthsValid(c) <= maxMonths
}

// SignatureString returns the TLS signature string corresponding to
// an X509 signature algorithm.
var signatureString = map[x509.SignatureAlgorithm]string{
	x509.UnknownSignatureAlgorithm: "Unknown Signature",
	x509.MD2WithRSA:                "MD2WithRSA",
	x509.MD5WithRSA:                "MD5WithRSA",
	x509.SHA1WithRSA:               "SHA1WithRSA",
	x509.SHA256WithRSA:             "SHA256WithRSA",
	x509.SHA384WithRSA:             "SHA384WithRSA",
	x509.SHA512WithRSA:             "SHA512WithRSA",
	x509.SHA256WithRSAPSS:          "SHA256WithRSAPSS",
	x509.SHA384WithRSAPSS:          "SHA384WithRSAPSS",
	x509.SHA512WithRSAPSS:          "SHA512WithRSAPSS",
	x509.DSAWithSHA1:               "DSAWithSHA1",
	x509.DSAWithSHA256:             "DSAWithSHA256",
	x509.ECDSAWithSHA1:             "ECDSAWithSHA1",
	x509.ECDSAWithSHA256:           "ECDSAWithSHA256",
	x509.ECDSAWithSHA384:           "ECDSAWithSHA384",
	x509.ECDSAWithSHA512:           "ECDSAWithSHA512",
	x509.PureEd25519:               "PureEd25519",
}

// SignatureString returns the TLS signature string corresponding to
// an X509 signature algorithm.
func SignatureString(alg x509.SignatureAlgorithm) string {
	if s, ok := signatureString[alg]; ok {
		return s
	}
	return "Unknown Signature"
}

// HashAlgoString returns the hash algorithm name contains in the signature
// method.
var hashAlgoString = map[x509.SignatureAlgorithm]string{
	x509.UnknownSignatureAlgorithm: "Unknown Hash Algorithm",
	x509.MD2WithRSA:                "MD2",
	x509.MD5WithRSA:                "MD5",
	x509.SHA1WithRSA:               "SHA1",
	x509.SHA256WithRSA:             "SHA256",
	x509.SHA384WithRSA:             "SHA384",
	x509.SHA512WithRSA:             "SHA512",
	x509.SHA256WithRSAPSS:          "SHA256",
	x509.SHA384WithRSAPSS:          "SHA384",
	x509.SHA512WithRSAPSS:          "SHA512",
	x509.DSAWithSHA1:               "SHA1",
	x509.DSAWithSHA256:             "SHA256",
	x509.ECDSAWithSHA1:             "SHA1",
	x509.ECDSAWithSHA256:           "SHA256",
	x509.ECDSAWithSHA384:           "SHA384",
	x509.ECDSAWithSHA512:           "SHA512",
	x509.PureEd25519:               "SHA512", // per x509 docs Ed25519 uses SHA-512 internally
}

// HashAlgoString returns the hash algorithm name contains in the signature
// method.
func HashAlgoString(alg x509.SignatureAlgorithm) string {
	if s, ok := hashAlgoString[alg]; ok {
		return s
	}
	return "Unknown Hash Algorithm"
}

// StringTLSVersion returns underlying enum values from human names for TLS
// versions, defaults to current golang default of TLS 1.0.
func StringTLSVersion(version string) uint16 {
	switch version {
	case "1.3":
		return tls.VersionTLS13
	case "1.2":
		return tls.VersionTLS12
	case "1.1":
		return tls.VersionTLS11
	case "1.0":
		return tls.VersionTLS10
	default:
		// Default to Go's historical default of TLS 1.0 for unknown values
		return tls.VersionTLS10
	}
}

// EncodeCertificatesPEM encodes a number of x509 certificates to PEM.
func EncodeCertificatesPEM(certs []*x509.Certificate) []byte {
	var buffer bytes.Buffer
	for _, cert := range certs {
		if err := pem.Encode(&buffer, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}); err != nil {
			return nil
		}
	}

	return buffer.Bytes()
}

// EncodeCertificatePEM encodes a single x509 certificates to PEM.
func EncodeCertificatePEM(cert *x509.Certificate) []byte {
	return EncodeCertificatesPEM([]*x509.Certificate{cert})
}

// ParseCertificatesPEM parses a sequence of PEM-encoded certificate and returns them,
// can handle PEM encoded PKCS #7 structures.
func ParseCertificatesPEM(certsPEM []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	var err error
	certsPEM = bytes.TrimSpace(certsPEM)
	for len(certsPEM) > 0 {
		var cert []*x509.Certificate
		cert, certsPEM, err = ParseOneCertificateFromPEM(certsPEM)
		if err != nil {
			return nil, certerr.ParsingError(certerr.ErrorSourceCertificate, err)
		} else if cert == nil {
			break
		}

		certs = append(certs, cert...)
	}
	if len(certsPEM) > 0 {
		return nil, certerr.DecodeError(
			certerr.ErrorSourceCertificate,
			errors.New("trailing data at end of certificate"),
		)
	}
	return certs, nil
}

// ParseCertificatesDER parses a DER encoding of a certificate object and possibly private key,
// either PKCS #7, PKCS #12, or raw x509.
func ParseCertificatesDER(certsDER []byte, password string) ([]*x509.Certificate, crypto.Signer, error) {
	certsDER = bytes.TrimSpace(certsDER)

	// First, try PKCS #7
	if pkcs7data, err7 := pkcs7.ParsePKCS7(certsDER); err7 == nil {
		if pkcs7data.ContentInfo != "SignedData" {
			return nil, nil, certerr.DecodeError(
				certerr.ErrorSourceCertificate,
				errors.New("can only extract certificates from signed data content info"),
			)
		}
		certs := pkcs7data.Content.SignedData.Certificates
		if certs == nil {
			return nil, nil, certerr.DecodeError(certerr.ErrorSourceCertificate, errors.New("no certificates decoded"))
		}
		return certs, nil, nil
	}

	// Next, try PKCS #12
	if pkcs12data, cert, err12 := pkcs12.Decode(certsDER, password); err12 == nil {
		signer, ok := pkcs12data.(crypto.Signer)
		if !ok {
			return nil, nil, certerr.DecodeError(
				certerr.ErrorSourcePrivateKey,
				errors.New("PKCS12 data does not contain a private key"),
			)
		}
		return []*x509.Certificate{cert}, signer, nil
	}

	// Finally, attempt to parse raw X.509 certificates
	certs, err := x509.ParseCertificates(certsDER)
	if err != nil {
		return nil, nil, certerr.DecodeError(certerr.ErrorSourceCertificate, err)
	}
	return certs, nil, nil
}

// ParseSelfSignedCertificatePEM parses a PEM-encoded certificate and check if it is self-signed.
func ParseSelfSignedCertificatePEM(certPEM []byte) (*x509.Certificate, error) {
	cert, err := ParseCertificatePEM(certPEM)
	if err != nil {
		return nil, err
	}

	err = cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
	if err != nil {
		return nil, certerr.VerifyError(certerr.ErrorSourceCertificate, err)
	}
	return cert, nil
}

// ParseCertificatePEM parses and returns a PEM-encoded certificate,
// can handle PEM encoded PKCS #7 structures.
func ParseCertificatePEM(certPEM []byte) (*x509.Certificate, error) {
	certPEM = bytes.TrimSpace(certPEM)
	certs, rest, err := ParseOneCertificateFromPEM(certPEM)
	if err != nil {
		return nil, certerr.ParsingError(certerr.ErrorSourceCertificate, err)
	}
	if certs == nil {
		return nil, certerr.DecodeError(certerr.ErrorSourceCertificate, errors.New("no certificate decoded"))
	}
	if len(rest) > 0 {
		return nil, certerr.ParsingError(
			certerr.ErrorSourceCertificate,
			errors.New("the PEM file should contain only one object"),
		)
	}
	if len(certs) > 1 {
		return nil, certerr.ParsingError(
			certerr.ErrorSourceCertificate,
			errors.New("the PKCS7 object in the PEM file should contain only one certificate"),
		)
	}
	return certs[0], nil
}

// ParseOneCertificateFromPEM attempts to parse one PEM encoded certificate object,
// either a raw x509 certificate or a PKCS #7 structure possibly containing
// multiple certificates, from the top of certsPEM, which itself may
// contain multiple PEM encoded certificate objects.
func ParseOneCertificateFromPEM(certsPEM []byte) ([]*x509.Certificate, []byte, error) {
	block, rest := pem.Decode(certsPEM)
	if block == nil {
		return nil, rest, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		pkcs7data, err2 := pkcs7.ParsePKCS7(block.Bytes)
		if err2 != nil {
			return nil, rest, err
		}
		if pkcs7data.ContentInfo != "SignedData" {
			return nil, rest, errors.New("only PKCS #7 Signed Data Content Info supported for certificate parsing")
		}
		certs := pkcs7data.Content.SignedData.Certificates
		if certs == nil {
			return nil, rest, errors.New("PKCS #7 structure contains no certificates")
		}
		return certs, rest, nil
	}
	var certs = []*x509.Certificate{cert}
	return certs, rest, nil
}

// LoadPEMCertPool loads a pool of PEM certificates from file.
func LoadPEMCertPool(certsFile string) (*x509.CertPool, error) {
	if certsFile == "" {
		return nil, nil //nolint:nilnil // no CA file provided -> treat as no pool and no error
	}
	pemCerts, err := os.ReadFile(certsFile)
	if err != nil {
		return nil, err
	}

	return PEMToCertPool(pemCerts)
}

// PEMToCertPool concerts PEM certificates to a CertPool.
func PEMToCertPool(pemCerts []byte) (*x509.CertPool, error) {
	if len(pemCerts) == 0 {
		return nil, nil //nolint:nilnil // empty input means no pool needed
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemCerts) {
		return nil, certerr.LoadingError(certerr.ErrorSourceCertificate, errors.New("failed to load cert pool"))
	}

	return certPool, nil
}

// ParsePrivateKeyPEM parses and returns a PEM-encoded private
// key. The private key may be either an unencrypted PKCS#8, PKCS#1,
// or elliptic private key.
func ParsePrivateKeyPEM(keyPEM []byte) (crypto.Signer, error) {
	return ParsePrivateKeyPEMWithPassword(keyPEM, nil)
}

// ParsePrivateKeyPEMWithPassword parses and returns a PEM-encoded private
// key. The private key may be a potentially encrypted PKCS#8, PKCS#1,
// or elliptic private key.
func ParsePrivateKeyPEMWithPassword(keyPEM []byte, password []byte) (crypto.Signer, error) {
	keyDER, err := GetKeyDERFromPEM(keyPEM, password)
	if err != nil {
		return nil, err
	}

	return ParsePrivateKeyDER(keyDER)
}

// GetKeyDERFromPEM parses a PEM-encoded private key and returns DER-format key bytes.
func GetKeyDERFromPEM(in []byte, password []byte) ([]byte, error) {
	// Ignore any EC PARAMETERS blocks when looking for a key (openssl includes
	// them by default).
	var keyDER *pem.Block
	for {
		keyDER, in = pem.Decode(in)
		if keyDER == nil || keyDER.Type != "EC PARAMETERS" {
			break
		}
	}
	if keyDER == nil {
		return nil, certerr.DecodeError(certerr.ErrorSourcePrivateKey, errors.New("failed to decode private key"))
	}
	if procType, ok := keyDER.Headers["Proc-Type"]; ok && strings.Contains(procType, "ENCRYPTED") {
		if password != nil {
			return x509.DecryptPEMBlock(keyDER, password)
		}
		return nil, certerr.DecodeError(certerr.ErrorSourcePrivateKey, certerr.ErrEncryptedPrivateKey)
	}
	return keyDER.Bytes, nil
}

// ParseCSR parses a PEM- or DER-encoded PKCS #10 certificate signing request.
func ParseCSR(in []byte) (*x509.CertificateRequest, []byte, error) {
	in = bytes.TrimSpace(in)
	p, rest := pem.Decode(in)
	if p == nil {
		csr, err := x509.ParseCertificateRequest(in)
		if err != nil {
			return nil, rest, certerr.ParsingError(certerr.ErrorSourceCSR, err)
		}
		if sigErr := csr.CheckSignature(); sigErr != nil {
			return nil, rest, certerr.VerifyError(certerr.ErrorSourceCSR, sigErr)
		}
		return csr, rest, nil
	}

	if p.Type != "NEW CERTIFICATE REQUEST" && p.Type != "CERTIFICATE REQUEST" {
		return nil, rest, certerr.ParsingError(
			certerr.ErrorSourceCSR,
			certerr.ErrInvalidPEMType(p.Type, "NEW CERTIFICATE REQUEST", "CERTIFICATE REQUEST"),
		)
	}

	csr, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return nil, rest, certerr.ParsingError(certerr.ErrorSourceCSR, err)
	}
	if sigErr := csr.CheckSignature(); sigErr != nil {
		return nil, rest, certerr.VerifyError(certerr.ErrorSourceCSR, sigErr)
	}
	return csr, rest, nil
}

// ParseCSRPEM parses a PEM-encoded certificate signing request.
// It does not check the signature. This is useful for dumping data from a CSR
// locally.
func ParseCSRPEM(csrPEM []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(csrPEM)
	if block == nil {
		return nil, certerr.DecodeError(certerr.ErrorSourceCSR, errors.New("PEM block is empty"))
	}
	csrObject, err := x509.ParseCertificateRequest(block.Bytes)

	if err != nil {
		return nil, certerr.ParsingError(certerr.ErrorSourceCSR, err)
	}

	return csrObject, nil
}

// SignerAlgo returns an X.509 signature algorithm from a crypto.Signer.
func SignerAlgo(priv crypto.Signer) x509.SignatureAlgorithm {
	const (
		rsaBits2048 = 2048
		rsaBits3072 = 3072
		rsaBits4096 = 4096
	)
	switch pub := priv.Public().(type) {
	case *rsa.PublicKey:
		bitLength := pub.N.BitLen()
		switch {
		case bitLength >= rsaBits4096:
			return x509.SHA512WithRSA
		case bitLength >= rsaBits3072:
			return x509.SHA384WithRSA
		case bitLength >= rsaBits2048:
			return x509.SHA256WithRSA
		default:
			return x509.SHA1WithRSA
		}
	case *ecdsa.PublicKey:
		switch pub.Curve {
		case elliptic.P521():
			return x509.ECDSAWithSHA512
		case elliptic.P384():
			return x509.ECDSAWithSHA384
		case elliptic.P256():
			return x509.ECDSAWithSHA256
		default:
			return x509.ECDSAWithSHA1
		}
	default:
		return x509.UnknownSignatureAlgorithm
	}
}

// LoadClientCertificate load key/certificate from pem files.
func LoadClientCertificate(certFile string, keyFile string) (*tls.Certificate, error) {
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, certerr.LoadingError(certerr.ErrorSourceKeypair, err)
		}
		return &cert, nil
	}
	return nil, nil //nolint:nilnil // absence of client cert is not an error
}

// CreateTLSConfig creates a tls.Config object from certs and roots.
func CreateTLSConfig(remoteCAs *x509.CertPool, cert *tls.Certificate) *tls.Config {
	var certs []tls.Certificate
	if cert != nil {
		certs = []tls.Certificate{*cert}
	}
	return &tls.Config{
		Certificates: certs,
		RootCAs:      remoteCAs,
		MinVersion:   tls.VersionTLS12, // secure default
	}
}

// SerializeSCTList serializes a list of SCTs.
func SerializeSCTList(sctList []ct.SignedCertificateTimestamp) ([]byte, error) {
	list := ctx509.SignedCertificateTimestampList{}
	for _, sct := range sctList {
		sctBytes, err := cttls.Marshal(sct)
		if err != nil {
			return nil, err
		}
		list.SCTList = append(list.SCTList, ctx509.SerializedSCT{Val: sctBytes})
	}
	return cttls.Marshal(list)
}

// DeserializeSCTList deserializes a list of SCTs.
func DeserializeSCTList(serializedSCTList []byte) ([]ct.SignedCertificateTimestamp, error) {
	var sctList ctx509.SignedCertificateTimestampList
	rest, err := cttls.Unmarshal(serializedSCTList, &sctList)
	if err != nil {
		return nil, err
	}
	if len(rest) != 0 {
		return nil, certerr.ParsingError(
			certerr.ErrorSourceSCTList,
			errors.New("serialized SCT list contained trailing garbage"),
		)
	}

	list := make([]ct.SignedCertificateTimestamp, len(sctList.SCTList))
	for i, serializedSCT := range sctList.SCTList {
		var sct ct.SignedCertificateTimestamp
		rest2, err2 := cttls.Unmarshal(serializedSCT.Val, &sct)
		if err2 != nil {
			return nil, err2
		}
		if len(rest2) != 0 {
			return nil, certerr.ParsingError(
				certerr.ErrorSourceSCTList,
				errors.New("serialized SCT list contained trailing garbage"),
			)
		}
		list[i] = sct
	}
	return list, nil
}

// SCTListFromOCSPResponse extracts the SCTList from an ocsp.Response,
// returning an empty list if the SCT extension was not found or could not be
// unmarshalled.
func SCTListFromOCSPResponse(response *ocsp.Response) ([]ct.SignedCertificateTimestamp, error) {
	// This loop finds the SCTListExtension in the OCSP response.
	var sctListExtension, ext pkix.Extension
	for _, ext = range response.Extensions {
		// sctExtOid is the ObjectIdentifier of a Signed Certificate Timestamp.
		sctExtOid := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 4, 5}
		if ext.Id.Equal(sctExtOid) {
			sctListExtension = ext
			break
		}
	}

	// This code block extracts the sctList from the SCT extension.
	var sctList []ct.SignedCertificateTimestamp
	var err error
	if numBytes := len(sctListExtension.Value); numBytes != 0 {
		var serializedSCTList []byte
		rest := make([]byte, numBytes)
		copy(rest, sctListExtension.Value)
		for len(rest) != 0 {
			rest, err = asn1.Unmarshal(rest, &serializedSCTList)
			if err != nil {
				return nil, certerr.ParsingError(certerr.ErrorSourceSCTList, err)
			}
		}
		sctList, err = DeserializeSCTList(serializedSCTList)
	}
	return sctList, err
}

// ReadBytes reads a []byte either from a file or an environment variable.
// If valFile has a prefix of 'env:', the []byte is read from the environment
// using the subsequent name. If the prefix is 'file:' the []byte is read from
// the subsequent file. If no prefix is provided, valFile is assumed to be a
// file path.
func ReadBytes(valFile string) ([]byte, error) {
	prefix, rest, found := strings.Cut(valFile, ":")
	if !found {
		return os.ReadFile(valFile)
	}
	switch prefix {
	case "env":
		return []byte(os.Getenv(rest)), nil
	case "file":
		return os.ReadFile(rest)
	default:
		return nil, fmt.Errorf("unknown prefix: %s", prefix)
	}
}
