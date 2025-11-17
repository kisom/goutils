package twofactor

import (
	"io"
	"testing"
)

func TestHOTPString(t *testing.T) {
	hotp := NewHOTP(nil, 0, 6)
	hotpString := otpString(hotp)
	if hotpString != "OATH-HOTP, 6" {
		t.Fatal("twofactor: invalid OTP string")
	}
}

// This test generates a new OTP, outputs the URL for that OTP,
// and attempts to parse that URL. It verifies that the two OTPs
// are the same, and that they produce the same output.
func TestURL(t *testing.T) {
	var ident = "testuser@foo"
	otp := NewHOTP(testKey, 0, 6)
	url := otp.URL("testuser@foo")
	otp2, id, err := FromURL(url)
	switch {
	case err != nil:
		t.Fatal("hotp: failed to parse HOTP URL\n")
	case id != ident:
		t.Logf("hotp: bad label\n")
		t.Logf("\texpected: %s\n", ident)
		t.Fatalf("\t  actual: %s\n", id)
	case otp2.Counter() != otp.Counter():
		t.Logf("hotp: OTP counters aren't synced\n")
		t.Logf("\toriginal: %d\n", otp.Counter())
		t.Fatalf("\t  second: %d\n", otp2.Counter())
	}

	code1 := otp.OTP()
	code2 := otp2.OTP()
	if code1 != code2 {
		t.Logf("hotp: mismatched OTPs\n")
		t.Logf("\texpected: %s\n", code1)
		t.Fatalf("\t  actual: %s\n", code2)
	}

	// There's not much we can do test the QR code, except to
	// ensure it doesn't fail.
	_, err = otp.QR(ident)
	if err != nil {
		t.Fatalf("hotp: failed to generate QR code PNG (%v)\n", err)
	}

	// This should fail because the maximum size of an alphanumeric
	// QR code with the lowest-level of error correction should
	// max out at 4296 bytes. 8k may be a bit overkill... but it
	// gets the job done. The value is read from the PRNG to
	// increase the likelihood that the returned data is
	// uncompressible.
	var tooBigIdent = make([]byte, 8192)
	_, err = io.ReadFull(PRNG, tooBigIdent)
	if err != nil {
		t.Fatalf("hotp: failed to read identity (%v)\n", err)
	} else if _, err = otp.QR(string(tooBigIdent)); err == nil {
		t.Fatal("hotp: QR code should fail to encode oversized URL")
	}
}

// This test makes sure we can generate codes for padded and non-padded
// entries.
func TestPaddedURL(t *testing.T) {
	var urlList = []string{
		"otpauth://hotp/?secret=ME",
		"otpauth://hotp/?secret=MEFR",
		"otpauth://hotp/?secret=MFRGG",
		"otpauth://hotp/?secret=MFRGGZA",
		"otpauth://hotp/?secret=a6mryljlbufszudtjdt42nh5by=======",
		"otpauth://hotp/?secret=a6mryljlbufszudtjdt42nh5by",
		"otpauth://hotp/?secret=a6mryljlbufszudtjdt42nh5by%3D%3D%3D%3D%3D%3D%3D",
	}
	var codeList = []string{
		"413198",
		"770938",
		"670717",
		"402378",
		"069864",
		"069864",
		"069864",
	}

	for i := range urlList {
		if o, id, err := FromURL(urlList[i]); err != nil {
			t.Log("hotp: URL should have parsed successfully (id=", id, ")")
			t.Logf("\turl was: %s\n", urlList[i])
			t.Fatalf("\t%s, %s\n", o.OTP(), id)
		} else {
			code2 := o.OTP()
			if code2 != codeList[i] {
				t.Logf("hotp: mismatched OTPs\n")
				t.Logf("\texpected: %s\n", codeList[i])
				t.Fatalf("\t  actual: %s\n", code2)
			}
		}
	}
}

// This test attempts a variety of invalid urls against the parser
// to ensure they fail.
func TestBadURL(t *testing.T) {
	var urlList = []string{
		"http://google.com",
		"",
		"-",
		"foo",
		"otpauth:/foo/bar/baz",
		"://",
		"otpauth://hotp/?digits=",
		"otpauth://hotp/?secret=MFRGGZDF&digits=ABCD",
		"otpauth://hotp/?secret=MFRGGZDF&counter=ABCD",
	}

	for i := range urlList {
		if _, _, err := FromURL(urlList[i]); err == nil {
			t.Log("hotp: URL should not have parsed successfully")
			t.Fatalf("\turl was: %s\n", urlList[i])
		}
	}
}
