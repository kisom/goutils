package mru

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestBasicCacheEviction(t *testing.T) {
	mock := clock.NewMock()
	c := New(2)
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

	mock.Add(time.Second)
	v, ok := c.Get("owl")
	if !ok {
		t.Fatal("store should have an entry for owl")
	}
	if err := c.ConsistencyCheck(); err != nil {
		t.Fatal(err)
	}

	itm, ok := v.(int)
	if !ok {
		t.Fatalf("stored item is not an int; have %T", v)
	}
	if err := c.ConsistencyCheck(); err != nil {
		t.Fatal(err)
	}

	if itm != 2 {
		t.Fatalf("stored item should be 2, have %d", itm)
	}

	mock.Add(time.Second)
	c.Store("elk", 4)
	if err := c.ConsistencyCheck(); err != nil {
		t.Fatal(err)
	}

	if !c.Has("elk") {
		t.Fatal("store should contain an entry for 'elk'")
	}

	if !c.Has("owl") {
		t.Fatal("store should contain an entry for 'owl'")
	}

	if c.Has("goat") {
		t.Fatal("store should not contain an entry for 'goat'")
	}
}
