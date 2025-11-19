package certgen

import (
	"encoding/asn1"
)

var (
	oidEd25519 = asn1.ObjectIdentifier{1, 3, 101, 110}
)

func GenerateKey() {}
