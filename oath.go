package twofactor

import (
	"crypto"
	"crypto/hmac"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"net/url"
)

const defaultSize = 6

// oath provides a baseline struct for the two OATH algorithms.
type oath struct {
	key      []byte
	counter  uint64
	size     int
	hash     func() hash.Hash
	algo     crypto.Hash
	provider string
}

// truncate contains the DT function from the RFC; this is used to
// deterministically select a sequence of 4 bytes from the HMAC
// counter hash.
func truncate(in []byte) int64 {
	offset := int(in[len(in)-1] & 0xF)
	p := in[offset : offset+4]
	var binCode int32
	binCode = int32((p[0] & 0x7f)) << 24
	binCode += int32((p[1] & 0xff)) << 16
	binCode += int32((p[2] & 0xff)) << 8
	binCode += int32((p[3] & 0xff))
	return int64(binCode) & 0x7FFFFFFF
}

func (o oath) Size() int {
	return o.size
}

func (o oath) Counter() uint64 {
	return o.counter
}

func (o oath) SetCounter(counter uint64) {
	o.counter = counter
}

func (o oath) Key() []byte {
	return o.key[:]
}

func (o oath) Hash() func() hash.Hash {
	return o.hash
}

func (o oath) URL(t Type, label string) string {
	secret := base32.StdEncoding.EncodeToString(o.key)
	u := url.URL{}
	v := url.Values{}
	u.Scheme = "otpauth"
	switch t {
	case OATH_HOTP:
		u.Host = "hotp"
	case OATH_TOTP:
		u.Host = "totp"
	}
	u.Path = label
	v.Add("secret", secret)
	if o.Counter() != 0 && t == OATH_HOTP {
		v.Add("counter", fmt.Sprintf("%d", o.Counter()))
	}
	if o.Size() != defaultSize {
		v.Add("digits", fmt.Sprintf("%d", o.Size()))
	}

	switch {
	case o.algo == crypto.SHA256:
		v.Add("algorithm", "SHA256")
	case o.algo == crypto.SHA512:
		v.Add("algorithm", "SHA512")
	}

	if o.provider != "" {
		v.Add("provider", o.provider)
	}

	u.RawQuery = v.Encode()
	return u.String()

}

func (o oath) QR(label string) ([]byte, error) {
	return nil, nil
}

var digits = []int{
	0:  1,
	1:  10,
	2:  100,
	3:  1000,
	4:  10000,
	5:  100000,
	6:  1000000,
	7:  10000000,
	8:  100000000,
	9:  1000000000,
	10: 10000000000,
}

// The top-level type should provide a counter; for example, HOTP
// will provide the counter directly while TOTP will provide the
// time-stepped counter.
func (o oath) OTP(counter uint64) string {
	var ctr [8]byte
	binary.BigEndian.PutUint64(ctr[:], counter)

	var mod int = 1
	if len(digits) > o.size {
		for i := 1; i <= o.size; i++ {
			mod *= 10
		}
	} else {
		mod = digits[o.size]
	}

	h := hmac.New(o.hash, o.key)
	h.Write(ctr[:])
	dt := truncate(h.Sum(nil))
	dt = dt % int64(mod)
	fmtStr := fmt.Sprintf("%%%dd", o.size)
	return fmt.Sprintf(fmtStr, dt)
}
