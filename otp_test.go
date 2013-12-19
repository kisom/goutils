package twofactor

import "fmt"
import "testing"

func TestHOTPString(t *testing.T) {
	hotp := NewHOTP(nil, 0, 6)
	hotpString := OTPString(hotp)
	if hotpString != "OATH-HOTP, 6" {
		fmt.Println("twofactor: invalid OTP string")
		t.FailNow()
	}
}
