package twofactor_test

import (
	"encoding/base32"
	"math/rand"
	"strings"
	"testing"

	"git.wntrmute.dev/kyle/goutils/twofactor"
)

const letters = "1234567890!@#$%^&*()abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString() string {
	b := make([]byte, rand.Intn(len(letters)))
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return base32.StdEncoding.EncodeToString(b)
}

func TestPadding(t *testing.T) {
	for range 300 {
		b := randString()
		origEncoding := b
		modEncoding := strings.ReplaceAll(b, "=", "")
		str, err := base32.StdEncoding.DecodeString(origEncoding)
		if err != nil {
			t.Fatal("Can't decode: ", b)
		}

		paddedEncoding := twofactor.Pad(modEncoding)
		if origEncoding != paddedEncoding {
			t.Log("Padding failed:")
			t.Logf("Expected: '%s'", origEncoding)
			t.Fatalf("Got: '%s'", paddedEncoding)
		} else {
			var mstr []byte
			mstr, err = base32.StdEncoding.DecodeString(paddedEncoding)
			if err != nil {
				t.Fatal("Can't decode: ", paddedEncoding)
			}

			if string(mstr) != string(str) {
				t.Log("Re-padding failed:")
				t.Logf("Expected: '%s'", str)
				t.Fatalf("Got: '%s'", mstr)
			}
		}
	}
}
