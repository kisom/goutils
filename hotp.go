package twofactor

import (
	"crypto"
	"crypto/sha1"
	"encoding/base32"
	"io"
	"net/url"
	"strconv"
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

func GenerateGoogleHOTP() *HOTP {
	key := make([]byte, sha1.Size)
	if _, err := io.ReadFull(PRNG, key); err != nil {
		return nil
	}
	return NewHOTP(key, 0, 6)
}

func hotpFromURL(u *url.URL) (*HOTP, string, error) {
	label := u.Path[1:]
	v := u.Query()

	secret := v.Get("secret")
	if secret == "" {
		return nil, "", ErrInvalidURL
	}

	var digits = 6
	if sdigit := v.Get("digits"); sdigit != "" {
		tmpDigits, err := strconv.ParseInt(sdigit, 10, 8)
		if err != nil {
			return nil, "", err
		}
		digits = int(tmpDigits)
	}

	var counter uint64 = 0
	if scounter := v.Get("counter"); scounter != "" {
		var err error
		counter, err = strconv.ParseUint(scounter, 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, "", err
	}
	otp := NewHOTP(key, counter, digits)
	return otp, label, nil
}

func (otp *HOTP) QR(label string) ([]byte, error) {
	return otp.oath.QR(otp.Type(), label)
}
