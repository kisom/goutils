package certlib_test

import (
	"testing"

	"git.wntrmute.dev/kyle/goutils/certlib"
)

var (
	testCert1 = "testdata/cert1.pem"
	testCert2 = "testdata/cert2.pem"
	testPriv1 = "testdata/priv1.pem"
	testPriv2 = "testdata/priv2.pem"
)

type testCase struct {
	cert  string
	key   string
	match bool
}

var testCases = []testCase{
	{testCert1, testPriv1, true},
	{testCert2, testPriv2, true},
	{testCert1, testPriv2, false},
	{testCert2, testPriv1, false},
}

func TestMatchKeys(t *testing.T) {
	for i, tc := range testCases {
		cert, err := certlib.LoadCertificate(tc.cert)
		if err != nil {
			t.Fatalf("failed to load cert %d: %v", i, err)
		}

		priv, err := certlib.LoadPrivateKey(tc.key)
		if err != nil {
			t.Fatalf("failed to load key %d: %v", i, err)
		}

		ok, _ := certlib.MatchKeys(cert, priv)
		switch {
		case ok && !tc.match:
			t.Fatalf("case %d: cert %s/key %s should not match", i, tc.cert, tc.key)
		case !ok && tc.match:
			t.Fatalf("case %d: cert %s/key %s should match", i, tc.cert, tc.key)
		}
	}
}
