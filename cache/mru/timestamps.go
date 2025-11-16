package mru

import (
	"errors"
	"fmt"
	"io"
	"sort"
)

// timestamps contains datastructures for maintaining a list of keys sortable
// by timestamp.

type timestamp[K comparable] struct {
	t int64
	k K
}

type timestamps[K comparable] struct {
	ts  []timestamp[K]
	cap int
}

func newTimestamps[K comparable](icap int) *timestamps[K] {
	return &timestamps[K]{
		ts:  make([]timestamp[K], 0, icap),
		cap: icap,
	}
}

func (ts *timestamps[K]) K(i int) K {
	return ts.ts[i].k
}

func (ts *timestamps[K]) T(i int) int64 {
	return ts.ts[i].t
}

func (ts *timestamps[K]) Len() int {
	return len(ts.ts)
}

func (ts *timestamps[K]) Less(i, j int) bool {
	return ts.ts[i].t < ts.ts[j].t
}

func (ts *timestamps[K]) Swap(i, j int) {
	ts.ts[i], ts.ts[j] = ts.ts[j], ts.ts[i]
}

func (ts *timestamps[K]) Find(k K) (int, bool) {
	for i := range ts.ts {
		if ts.ts[i].k == k {
			return i, true
		}
	}
	return -1, false
}

func (ts *timestamps[K]) Update(k K, t int64) bool {
	i, ok := ts.Find(k)
	if !ok {
		ts.ts = append(ts.ts, timestamp[K]{t, k})
		sort.Sort(ts)
		return false
	}

	ts.ts[i].t = t
	sort.Sort(ts)
	return true
}

func (ts *timestamps[K]) ConsistencyCheck() error {
	if !sort.IsSorted(ts) {
		return errors.New("mru: timestamps are not sorted")
	}

	keys := map[K]bool{}
	for i := range ts.ts {
		if keys[ts.ts[i].k] {
			return fmt.Errorf("duplicate key %v detected", ts.ts[i].k)
		}
		keys[ts.ts[i].k] = true
	}

	if len(keys) != len(ts.ts) {
		return fmt.Errorf("mru: timestamp contains %d duplicate keys",
			len(ts.ts)-len(keys))
	}

	return nil
}

func (ts *timestamps[K]) Delete(i int) {
	ts.ts = append(ts.ts[:i], ts.ts[i+1:]...)
}

func (ts *timestamps[K]) Dump(w io.Writer) {
	for i := range ts.ts {
		fmt.Fprintf(w, "%d: %v, %d\n", i, ts.K(i), ts.T(i))
	}
}
