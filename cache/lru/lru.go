// Package lru implements a Least Recently Used cache.
package lru

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/benbjohnson/clock"
)

type item[V any] struct {
	V      V
	access int64
}

// A Cache is a map that retains a limited number of items. It must be
// initialized with New, providing a maximum capacity for the cache.
// Only the least recently used items are retained.
type Cache[K comparable, V any] struct {
	store  map[K]*item[V]
	access *timestamps[K]
	cap    int
	clock  clock.Clock
	// All public methods that have the possibility of modifying the
	// cache should lock it.
	mtx *sync.Mutex
}

// New must be used to create a new Cache.
func New[K comparable, V any](icap int) *Cache[K, V] {
	return &Cache[K, V]{
		store:  map[K]*item[V]{},
		access: newTimestamps[K](icap),
		cap:    icap,
		clock:  clock.New(),
		mtx:    &sync.Mutex{},
	}
}

// StringKeyCache is a convenience wrapper for cache keyed by string.
type StringKeyCache[V any] struct {
	*Cache[string, V]
}

// NewStringKeyCache creates a new LRU cache keyed by string.
func NewStringKeyCache[V any](icap int) *StringKeyCache[V] {
	return &StringKeyCache[V]{Cache: New[string, V](icap)}
}

func (c *Cache[K, V]) lock() {
	c.mtx.Lock()
}

func (c *Cache[K, V]) unlock() {
	c.mtx.Unlock()
}

// Len returns the number of items currently in the cache.
func (c *Cache[K, V]) Len() int {
	return len(c.store)
}

// evict should remove the least-recently-used cache item.
func (c *Cache[K, V]) evict() {
	if c.access.Len() == 0 {
		return
	}

	k := c.access.K(0)
	c.evictKey(k)
}

// evictKey should remove the entry given by the key item.
func (c *Cache[K, V]) evictKey(k K) {
	delete(c.store, k)
	i, ok := c.access.Find(k)
	if !ok {
		return
	}

	c.access.Delete(i)
}

func (c *Cache[K, V]) sanityCheck() {
	if len(c.store) != c.access.Len() {
		panic(fmt.Sprintf("LRU cache is out of sync; store len = %d, access len = %d",
			len(c.store), c.access.Len()))
	}
}

// ConsistencyCheck runs a series of checks to ensure that the cache's
// data structures are consistent. It is not normally required, and it
// is primarily used in testing.
func (c *Cache[K, V]) ConsistencyCheck() error {
	c.lock()
	defer c.unlock()
	if err := c.access.ConsistencyCheck(); err != nil {
		return err
	}

	if len(c.store) != c.access.Len() {
		return fmt.Errorf("lru: cache is out of sync; store len = %d, access len = %d",
			len(c.store), c.access.Len())
	}

	for i := range c.access.ts {
		itm, ok := c.store[c.access.K(i)]
		if !ok {
			return errors.New("lru: key in access is not in store")
		}

		if c.access.T(i) != itm.access {
			return fmt.Errorf("timestamps are out of sync (%d != %d)",
				itm.access, c.access.T(i))
		}
	}

	if !sort.IsSorted(c.access) {
		return errors.New("lru: timestamps aren't sorted")
	}

	return nil
}

// Store adds the value v to the cache under the k.
func (c *Cache[K, V]) Store(k K, v V) {
	c.lock()
	defer c.unlock()

	c.sanityCheck()

	if len(c.store) == c.cap {
		c.evict()
	}

	if _, ok := c.store[k]; ok {
		c.evictKey(k)
	}

	itm := &item[V]{
		V:      v,
		access: c.clock.Now().UnixNano(),
	}

	c.store[k] = itm
	c.access.Update(k, itm.access)
}

// Get returns the value stored in the cache. If the item isn't present,
// it will return false.
func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.lock()
	defer c.unlock()

	c.sanityCheck()

	itm, ok := c.store[k]
	if !ok {
		var zero V
		return zero, false
	}

	c.store[k].access = c.clock.Now().UnixNano()
	c.access.Update(k, itm.access)
	return itm.V, true
}

// Has returns true if the cache has an entry for k. It will not update
// the timestamp on the item.
func (c *Cache[K, V]) Has(k K) bool {
	// Don't need to lock as we don't modify anything.

	c.sanityCheck()

	_, ok := c.store[k]
	return ok
}
