package mru

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/benbjohnson/clock"
)

type item struct {
	V      any
	access int64
}

// A Cache is a map that retains a limited number of items. It must be
// initialized with New, providing a maximum capacity for the cache.
// Only the most recently used items are retained.
type Cache struct {
	store  map[string]*item
	access *timestamps
	cap    int
	clock  clock.Clock
	// All public methods that have the possibility of modifying the
	// cache should lock it.
	mtx *sync.Mutex
}

// New must be used to create a new Cache.
func New(icap int) *Cache {
	return &Cache{
		store:  map[string]*item{},
		access: newTimestamps(icap),
		cap:    icap,
		clock:  clock.New(),
		mtx:    &sync.Mutex{},
	}
}

func (c *Cache) lock() {
	c.mtx.Lock()
}

func (c *Cache) unlock() {
	c.mtx.Unlock()
}

// Len returns the number of items currently in the cache.
func (c *Cache) Len() int {
	return len(c.store)
}

// evict should remove the least-recently-used cache item.
func (c *Cache) evict() {
	if c.access.Len() == 0 {
		return
	}

	k := c.access.K(0)
	c.evictKey(k)
}

// evictKey should remove the entry given by the key item.
func (c *Cache) evictKey(k string) {
	delete(c.store, k)
	i, ok := c.access.Find(k)
	if !ok {
		return
	}

	c.access.Delete(i)
}

func (c *Cache) sanityCheck() {
	if len(c.store) != c.access.Len() {
		panic(fmt.Sprintf("MRU cache is out of sync; store len = %d, access len = %d",
			len(c.store), c.access.Len()))
	}
}

// ConsistencyCheck runs a series of checks to ensure that the cache's
// data structures are consistent. It is not normally required, and it
// is primarily used in testing.
func (c *Cache) ConsistencyCheck() error {
	c.lock()
	defer c.unlock()
	if err := c.access.ConsistencyCheck(); err != nil {
		return err
	}

	if len(c.store) != c.access.Len() {
		return fmt.Errorf("mru: cache is out of sync; store len = %d, access len = %d",
			len(c.store), c.access.Len())
	}

	for i := range c.access.ts {
		itm, ok := c.store[c.access.K(i)]
		if !ok {
			return errors.New("mru: key in access is not in store")
		}

		if c.access.T(i) != itm.access {
			return fmt.Errorf("timestamps are out of sync (%d != %d)",
				itm.access, c.access.T(i))
		}
	}

	if !sort.IsSorted(c.access) {
		return errors.New("mru: timestamps aren't sorted")
	}

	return nil
}

// Store adds the value v to the cache under the k.
func (c *Cache) Store(k string, v any) {
	c.lock()
	defer c.unlock()

	c.sanityCheck()

	if len(c.store) == c.cap {
		c.evict()
	}

	if _, ok := c.store[k]; ok {
		c.evictKey(k)
	}

	itm := &item{
		V:      v,
		access: c.clock.Now().UnixNano(),
	}

	c.store[k] = itm
	c.access.Update(k, itm.access)
}

// Get returns the value stored in the cache. If the item isn't present,
// it will return false.
func (c *Cache) Get(k string) (any, bool) {
	c.lock()
	defer c.unlock()

	c.sanityCheck()

	itm, ok := c.store[k]
	if !ok {
		return nil, false
	}

	c.store[k].access = c.clock.Now().UnixNano()
	c.access.Update(k, itm.access)
	return itm.V, true
}

// Has returns true if the cache has an entry for k. It will not update
// the timestamp on the item.
func (c *Cache) Has(k string) bool {
	// Don't need to lock as we don't modify anything.

	c.sanityCheck()

	_, ok := c.store[k]
	return ok
}
