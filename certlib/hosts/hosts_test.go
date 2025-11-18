package hosts_test

import (
	"testing"

	"git.wntrmute.dev/kyle/goutils/certlib/hosts"
)

type testCase struct {
	Host   string
	Target hosts.Target
}

var testCases = []testCase{
	{Host: "server-name", Target: hosts.Target{Host: "server-name", Port: 443}},
	{Host: "server-name:8443", Target: hosts.Target{Host: "server-name", Port: 8443}},
	{Host: "tls://server-name", Target: hosts.Target{Host: "server-name", Port: 443}},
	{Host: "https://server-name", Target: hosts.Target{Host: "server-name", Port: 443}},
	{Host: "https://server-name:8443", Target: hosts.Target{Host: "server-name", Port: 8443}},
	{Host: "tls://server-name:8443", Target: hosts.Target{Host: "server-name", Port: 8443}},
	{Host: "https://server-name/something/else", Target: hosts.Target{Host: "server-name", Port: 443}},
}

func TestParseHost(t *testing.T) {
	for i, tc := range testCases {
		target, err := hosts.ParseHost(tc.Host)
		if err != nil {
			t.Fatalf("test case %d: %s", i+1, err)
		}

		if target.Host != tc.Target.Host {
			t.Fatalf("test case %d: got host '%s', want host '%s'", i+1, target.Host, tc.Target.Host)
		}
	}
}
