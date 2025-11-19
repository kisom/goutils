package certgen

import "crypto/x509"

var keyUsageStrings = map[string]x509.KeyUsage{
	"signing":            x509.KeyUsageDigitalSignature,
	"digital signature":  x509.KeyUsageDigitalSignature,
	"content commitment": x509.KeyUsageContentCommitment,
	"key encipherment":   x509.KeyUsageKeyEncipherment,
	"key agreement":      x509.KeyUsageKeyAgreement,
	"data encipherment":  x509.KeyUsageDataEncipherment,
	"cert sign":          x509.KeyUsageCertSign,
	"crl sign":           x509.KeyUsageCRLSign,
	"encipher only":      x509.KeyUsageEncipherOnly,
	"decipher only":      x509.KeyUsageDecipherOnly,
}

var extKeyUsageStrings = map[string]x509.ExtKeyUsage{
	"any":              x509.ExtKeyUsageAny,
	"server auth":      x509.ExtKeyUsageServerAuth,
	"client auth":      x509.ExtKeyUsageClientAuth,
	"code signing":     x509.ExtKeyUsageCodeSigning,
	"email protection": x509.ExtKeyUsageEmailProtection,
	"s/mime":           x509.ExtKeyUsageEmailProtection,
	"ipsec end system": x509.ExtKeyUsageIPSECEndSystem,
	"ipsec tunnel":     x509.ExtKeyUsageIPSECTunnel,
	"ipsec user":       x509.ExtKeyUsageIPSECUser,
	"timestamping":     x509.ExtKeyUsageTimeStamping,
	"ocsp signing":     x509.ExtKeyUsageOCSPSigning,
	"microsoft sgc":    x509.ExtKeyUsageMicrosoftServerGatedCrypto,
	"netscape sgc":     x509.ExtKeyUsageNetscapeServerGatedCrypto,
}
