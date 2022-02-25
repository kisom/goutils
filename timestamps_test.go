package mru

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestTimestamps(t *testing.T) {
	ts := newTimestamps(3)
	mock := clock.NewMock()

	// raven
	ts.Update("raven", mock.Now().UnixNano())

	// raven, owl
	mock.Add(time.Millisecond)

	ts.Update("owl", mock.Now().UnixNano())

	// raven, owl, goat
	mock.Add(time.Second)
	ts.Update("goat", mock.Now().UnixNano())

	if err := ts.ConsistencyCheck(); err != nil {
		t.Fatal(err)
	}
	mock.Add(time.Millisecond)

	// raven, goat, owl
	ts.Update("owl", mock.Now().UnixNano())
	if err := ts.ConsistencyCheck(); err != nil {
		t.Fatal(err)
	}

	// at this point, the keys should be raven, goat, owl.
	if ts.K(0) != "raven" {
		t.Fatalf("first key should be raven, have %s", ts.K(0))
	}

	if ts.K(1) != "goat" {
		t.Fatalf("second key should be goat, have %s", ts.K(1))
	}

	if ts.K(2) != "owl" {
		t.Fatalf("third key should be owl, have %s", ts.K(2))
	}

}
