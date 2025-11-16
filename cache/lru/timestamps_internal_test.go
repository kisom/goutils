package lru

import (
    "testing"
    "time"

    "github.com/benbjohnson/clock"
)

// These tests validate timestamps ordering semantics for the LRU package.
// Note: The LRU timestamps are sorted with most-recent-first (descending by t).
func TestTimestamps(t *testing.T) {
    ts := newTimestamps[string](3)
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

    // make owl the most recent
    mock.Add(time.Millisecond)
    ts.Update("owl", mock.Now().UnixNano())
    if err := ts.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    // For LRU timestamps: most recent first. Expected order: owl, goat, raven.
    if ts.K(0) != "owl" {
        t.Fatalf("first key should be owl, have %s", ts.K(0))
    }

    if ts.K(1) != "goat" {
        t.Fatalf("second key should be goat, have %s", ts.K(1))
    }

    if ts.K(2) != "raven" {
        t.Fatalf("third key should be raven, have %s", ts.K(2))
    }
}
