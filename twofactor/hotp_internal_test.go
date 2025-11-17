package twofactor

import (
	"testing"
)

var testKey = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

var rfcHotpKey = []byte("12345678901234567890")
var rfcHotpExpected = []string{
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
func TestHotpRFC(t *testing.T) {
	otp := NewHOTP(rfcHotpKey, 0, 6)
	for i := range rfcHotpExpected {
		if otp.Counter() != uint64(i) {
			t.Fatalf("twofactor: invalid counter (should be %d, is %d",
				i, otp.Counter())
		}
		code := otp.OTP()
		if code == "" {
			t.Fatal("twofactor: failed to produce an OTP")
		} else if code != rfcHotpExpected[i] {
			t.Logf("twofactor: invalid OTP\n")
			t.Logf("\tExpected: %s\n", rfcHotpExpected[i])
			t.Logf("\t  Actual: %s\n", code)
			t.Fatalf("\t Counter: %d\n", otp.counter)
		}
	}
}

// This test uses a different key than the test cases in the RFC,
// but runs through the same test cases to ensure that they fail as
// expected.
func TestHotpBadRFC(t *testing.T) {
	otp := NewHOTP(testKey, 0, 6)
	for i := range rfcHotpExpected {
		code := otp.OTP()
		switch code {
		case "":
			t.Error("twofactor: failed to produce an OTP")
		case rfcHotpExpected[i]:
			t.Error("twofactor: should not have received a valid OTP")
		}
	}
}
