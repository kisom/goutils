package twofactor

import (
	"fmt"
	"testing"
)

var testKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

func newZeroHOTP() *HOTP {
	return NewHOTP(testKey, 0, 6)
}

var sha1Hmac = []byte{
	0x1f, 0x86, 0x98, 0x69, 0x0e,
	0x02, 0xca, 0x16, 0x61, 0x85,
	0x50, 0xef, 0x7f, 0x19, 0xda,
	0x8e, 0x94, 0x5b, 0x55, 0x5a,
}

var truncExpect int64 = 0x50ef7f19

// This test runs through the truncation example given in the RFC.
func TestTruncate(t *testing.T) {
	if result := truncate(sha1Hmac); result != truncExpect {
		fmt.Printf("hotp: expected truncate -> %d, saw %d\n",
			truncExpect, result)
		t.FailNow()
	}

	sha1Hmac[19]++
	if result := truncate(sha1Hmac); result == truncExpect {
		fmt.Println("hotp: expected truncation to fail")
		t.FailNow()
	}
}

var rfcKey = []byte("12345678901234567890")
var rfcExpected = []string{
	"755224",
	"287082",
	"359152",
	"969429",
	"338314",
	"254676",
	"287922",
	"162583",
	"399871",
	"520489",
}

// This test runs through the test cases presented in the RFC, and
// ensures that this implementation is in compliance.
func TestRFC(t *testing.T) {
	otp := NewHOTP(rfcKey, 0, 6)
	for i := 0; i < len(rfcExpected); i++ {
		if otp.Counter() != uint64(i) {
			fmt.Printf("hotp: invalid counter (should be %d, is %d",
				i, otp.Counter())
			t.FailNow()
		}
		code := otp.OTP()
		if code == "" {
			fmt.Printf("hotp: failed to produce an OTP\n")
			t.FailNow()
		} else if code != rfcExpected[i] {
			fmt.Printf("hotp: invalid OTP\n")
			fmt.Printf("\tExpected: %s\n", rfcExpected[i])
			fmt.Printf("\t  Actual: %s\n", code)
			fmt.Printf("\t Counter: %d\n", otp.counter)
			t.FailNow()
		}
	}
}

// This test uses a different key than the test cases in the RFC,
// but runs through the same test cases to ensure that they fail as
// expected.
func TestBadRFC(t *testing.T) {
	otp := NewHOTP(testKey, 0, 6)
	for i := 0; i < len(rfcExpected); i++ {
		code := otp.OTP()
		if code == "" {
			fmt.Printf("hotp: failed to produce an OTP\n")
			t.FailNow()
		} else if code == rfcExpected[i] {
			fmt.Printf("hotp: should not have received a valid OTP\n")
			t.FailNow()
		}
	}
}
