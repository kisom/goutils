package twofactor

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"hash"
	"io"
	"net/url"
	"strconv"
	"time"
)

type TOTP struct {
	*oath
	step uint64
}

func (otp *TOTP) Type() Type {
	return OATH_TOTP
}

func (otp *TOTP) otp(counter uint64) string {
	return otp.oath.OTP(counter)
}

func (otp *TOTP) OTP() string {
	return otp.otp(otp.OTPCounter())
}

func (otp *TOTP) URL(label string) string {
	return otp.oath.URL(otp.Type(), label)
}

func (otp *TOTP) SetProvider(provider string) {
	otp.provider = provider
}

func (otp *TOTP) otpCounter(t uint64) uint64 {
	return (t - otp.counter) / otp.step
}

func (otp *TOTP) OTPCounter() uint64 {
	return otp.otpCounter(uint64(time.Now().Unix()))
}

func NewTOTP(key []byte, start uint64, step uint64, digits int, algo crypto.Hash) *TOTP {
	h := hashFromAlgo(algo)
	if h == nil {
		return nil
	}

	return &TOTP{
		oath: &oath{
			key:     key,
			counter: start,
			size:    digits,
			hash:    h,
			algo:    algo,
		},
		step: step,
	}

}

func NewTOTPSHA1(key []byte, start uint64, step uint64, digits int) *TOTP {
	return NewTOTP(key, start, step, digits, crypto.SHA1)
}

func NewTOTPSHA256(key []byte, start uint64, step uint64, digits int) *TOTP {
	return NewTOTP(key, start, step, digits, crypto.SHA256)
}

func NewTOTPSHA512(key []byte, start uint64, step uint64, digits int) *TOTP {
	return NewTOTP(key, start, step, digits, crypto.SHA512)
}

func hashFromAlgo(algo crypto.Hash) func() hash.Hash {
	switch algo {
	case crypto.SHA1:
		return sha1.New
	case crypto.SHA256:
		return sha256.New
	case crypto.SHA512:
		return sha512.New
	}
	return nil
}

// GenerateGoogleTOTP produces a new TOTP token with the defaults expected by
// Google Authenticator.
func GenerateGoogleTOTP() *TOTP {
	key := make([]byte, sha1.Size)
	if _, err := io.ReadFull(PRNG, key); err != nil {
		return nil
	}
	return NewTOTP(key, 0, 30, 6, crypto.SHA1)
}

func totpFromURL(u *url.URL) (*TOTP, string, error) {
	label := u.Path[1:]
	v := u.Query()

	secret := v.Get("secret")
	if secret == "" {
		return nil, "", ErrInvalidURL
	}

	var algo = crypto.SHA1
	if algorithm := v.Get("algorithm"); algorithm != "" {
		switch {
		case algorithm == "SHA256":
			algo = crypto.SHA256
		case algorithm == "SHA512":
			algo = crypto.SHA512
		case algorithm != "SHA1":
			return nil, "", ErrInvalidAlgo
		}
	}

	var digits = 6
	if sdigit := v.Get("digits"); sdigit != "" {
		tmpDigits, err := strconv.ParseInt(sdigit, 10, 8)
		if err != nil {
			return nil, "", err
		}
		digits = int(tmpDigits)
	}

	var period uint64 = 30
	if speriod := v.Get("period"); speriod != "" {
		var err error
		period, err = strconv.ParseUint(speriod, 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, "", err
	}
	otp := NewTOTP(key, 0, period, digits, algo)
	return otp, label, nil
}

func (otp *TOTP) QR(label string) ([]byte, error) {
	return otp.oath.QR(otp.Type(), label)
}
