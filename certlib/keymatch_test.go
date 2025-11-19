package certlib

import "testing"

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
	{"testdata/cert1.pem", "testdata/priv1.pem", true},
	{"testdata/cert2.pem", "testdata/priv2.pem", true},
	{"testdata/cert1.pem", "testdata/priv2.pem", false},
	{"testdata/cert2.pem", "testdata/priv1.pem", false},
}

func TestMatchKeys(t *testing.T) {
	for i, tc := range testCases {
		cert, err := LoadCertificate(tc.cert)
		if err != nil {
			t.Fatalf("failed to load cert %d: %v", i, err)
		}

		priv, err := LoadPrivateKey(tc.key)
		if err != nil {
			t.Fatalf("failed to load key %d: %v", i, err)
		}

		ok, _ := MatchKeys(cert, priv)
		switch {
		case ok && !tc.match:
			t.Fatalf("case %d: cert %s/key %s should not match", i, tc.cert, tc.key)
		case !ok && tc.match:
			t.Fatalf("case %d: cert %s/key %s should match", i, tc.cert, tc.key)
		}
	}
}
