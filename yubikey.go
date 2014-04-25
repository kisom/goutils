package twofactor

// Implement YubiKey OTP and YubiKey HOTP.

import (
	"github.com/conformal/yubikey"
	"github.com/gokyle/twofactor/modhex"
	"hash"
)

// YubiKey is an implementation of the YubiKey hard token. Note
// that the internal counter only actually uses 32 bits.
type YubiKey struct {
	token   yubikey.Token
	counter uint64
	key     yubikey.Key
	public  []byte
}

// Public returns the public component of the token.
func (yk *YubiKey) Public() []byte {
	return yk.public[:]
}

// Counter returns the YubiKey's counter.
func (yk *YubiKey) Counter() uint64 {
	return yk.counter
}

// SetCounter sets the YubiKey's counter.
func (yk *YubiKey) SetCounter(counter uint64) {
	yk.counter = counter & 0xffffffff
}

// Key returns the YubiKey's secret key.
func (yk *YubiKey) Key() []byte {
	return yk.key[:]
}

// Size returns the length of the YubiKey's OTP output plus the length
// of the public identifier.
func (yk *YubiKey) Size() int {
	return yubikey.OTPSize + len(yk.public)
}

// OTP returns a new one-time password from the YubiKey.
func (yk *YubiKey) OTP() string {
	otp := yk.token.Generate(yk.key)
	if otp == nil {
		return ""
	}
	return modhex.StdEncoding.EncodeToString(otp.Bytes())
}

// Hash always returns nil, as the YubiKey tokens do not use a hash
// function.
func (yk *YubiKey) Hash() func() hash.Hash {
	return nil
}

// Type returns YUBIKEY.
func (yk *YubiKey) Type() Type {
	return YUBIKEY
}
