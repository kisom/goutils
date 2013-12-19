package twofactor

import (
	"crypto/rand"
	"fmt"
	"hash"
)

type Type uint

const (
	OATH_HOTP = iota
	OATH_TOTP
)

var PRNG = rand.Reader

// Type OTP represents a one-time password token -- whether a
// software taken (as in the case of Google Authenticator) or a
// hardware token (as in the case of a YubiKey).
type OTP interface {
	// Returns the current counter value; the meaning of the
	// returned value is algorithm-specific.
	Counter() uint64

	// Set the counter to a specific value.
	SetCounter(uint64)

	// the secret key contained in the OTP
	Key() []byte

	// generate a new OTP
	OTP() string

	// the output size of the OTP
	Size() int

	// the hash function used by the OTP
	Hash() func() hash.Hash

	// URL generates a Google Authenticator url (or perhaps some other url)
	URL(string) string

	// QR outputs a byte slice containing a PNG-encoded QR code
	// of the URL.
	QR(string) ([]byte, error)

	// Returns the type of this OTP.
	Type() Type
}

func OTPString(otp OTP) string {
	var typeName string
	switch otp.Type() {
	case OATH_HOTP:
		typeName = "OATH-HOTP"
	case OATH_TOTP:
		typeName = "OATH-TOTP"
	}
	return fmt.Sprintf("%s, %d", typeName, otp.Size())
}
