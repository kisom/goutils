package twofactor

import (
	"crypto"
	"crypto/sha1"
)

type HOTP struct {
	*oath
}

func (otp *HOTP) Type() Type {
	return OATH_HOTP
}

func NewHOTP(key []byte, counter uint64, digits int) *HOTP {
	return &HOTP{
		oath: &oath{
			key:     key,
			counter: counter,
			size:    digits,
			hash:    sha1.New,
			algo:    crypto.SHA1,
		},
	}
}

func (otp *HOTP) OTP() string {
	code := otp.oath.OTP(otp.counter)
	otp.counter++
	return code
}

func (otp *HOTP) URL(label string) string {
	return otp.oath.URL(otp.Type(), label)
}

func (otp *HOTP) SetProvider(provider string) {
	otp.provider = provider
}
