package certlib

import (
	"crypto/x509"
	"encoding/asn1"
	"github.com/davecgh/go-spew/spew"
	ct "github.com/google/certificate-transparency-go"
)

var sctExtension = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 11129, 2, 4, 2}

// SignedCertificateTimestampList is a list of signed certificate timestamps, from RFC6962 s3.3.
type SignedCertificateTimestampList struct {
	SCTList []ct.SignedCertificateTimestamp
}

func DumpSignedCertificateList(cert *x509.Certificate) ([]ct.SignedCertificateTimestamp, error) {
	// x := x509.SignedCertificateTimestampList{}
	var sctList []ct.SignedCertificateTimestamp

	for _, extension := range cert.Extensions {
		if extension.Id.Equal(sctExtension) {
			spew.Dump(extension)

			var rawSCT ct.SignedCertificateTimestamp
			_, err := asn1.Unmarshal(extension.Value, &rawSCT)
			if err != nil {
				return nil, err
			}

			sctList = append(sctList, rawSCT)
		}
	}

	return sctList, nil
}
