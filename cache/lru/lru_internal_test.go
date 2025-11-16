package lru

import (
    "testing"
    "time"

    "github.com/benbjohnson/clock"
)

// These tests mirror the MRU-style behavior present in this LRU package
// implementation (eviction removes the most-recently-used entry).
func TestBasicCacheEviction(t *testing.T) {
    mock := clock.NewMock()
    c := NewStringKeyCache[int](2)
    c.clock = mock

    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    if c.Len() != 0 {
        t.Fatal("cache should have size 0")
    }

    c.evict()
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    c.Store("raven", 1)
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    if len(c.store) != 1 {
        t.Fatalf("store should have length=1, have length=%d", len(c.store))
    }

    mock.Add(time.Second)
    c.Store("owl", 2)
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    if len(c.store) != 2 {
        t.Fatalf("store should have length=2, have length=%d", len(c.store))
    }

    mock.Add(time.Second)
    c.Store("goat", 3)
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    if len(c.store) != 2 {
        t.Fatalf("store should have length=2, have length=%d", len(c.store))
    }

    // Since this implementation evicts the most-recently-used item, inserting
    // "goat" when full evicts "owl" (the most recent at that time).
    mock.Add(time.Second)
    if _, ok := c.Get("owl"); ok {
        t.Fatal("store should not have an entry for owl (MRU-evicted)")
    }
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    mock.Add(time.Second)
    c.Store("elk", 4)
    if err := c.ConsistencyCheck(); err != nil {
        t.Fatal(err)
    }

    if !c.Has("elk") {
        t.Fatal("store should contain an entry for 'elk'")
    }

    // Before storing elk, keys were: raven (older), goat (newer). Evict MRU -> goat.
    if !c.Has("raven") {
        t.Fatal("store should contain an entry for 'raven'")
    }

    if c.Has("goat") {
        t.Fatal("store should not contain an entry for 'goat'")
    }
}
