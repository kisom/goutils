package mru

import (
	"errors"
	"fmt"
	"io"
	"sort"
)

// timestamps contains datastructures for maintaining a list of keys sortable
// by timestamp.

type timestamp struct {
	t int64
	k string
}

type timestamps struct {
	ts  []timestamp
	cap int
}

func newTimestamps(icap int) *timestamps {
	return &timestamps{
		ts:  make([]timestamp, 0, icap),
		cap: icap,
	}
}

func (ts *timestamps) K(i int) string {
	return ts.ts[i].k
}

func (ts *timestamps) T(i int) int64 {
	return ts.ts[i].t
}

func (ts *timestamps) Len() int {
	return len(ts.ts)
}

func (ts *timestamps) Less(i, j int) bool {
	return ts.ts[i].t < ts.ts[j].t
}

func (ts *timestamps) Swap(i, j int) {
	ts.ts[i], ts.ts[j] = ts.ts[j], ts.ts[i]
}

func (ts *timestamps) Find(k string) (int, bool) {
	for i := range len(ts.ts) {
		if ts.ts[i].k == k {
			return i, true
		}
	}
	return -1, false
}

func (ts *timestamps) Update(k string, t int64) bool {
	i, ok := ts.Find(k)
	if !ok {
		ts.ts = append(ts.ts, timestamp{t, k})
		sort.Sort(ts)
		return false
	}

	ts.ts[i].t = t
	sort.Sort(ts)
	return true
}

func (ts *timestamps) ConsistencyCheck() error {
	if !sort.IsSorted(ts) {
		return errors.New("mru: timestamps are not sorted")
	}

	keys := map[string]bool{}
	for i := range ts.ts {
		if keys[ts.ts[i].k] {
			return fmt.Errorf("duplicate key %s detected", ts.ts[i].k)
		}
		keys[ts.ts[i].k] = true
	}

	if len(keys) != len(ts.ts) {
		return fmt.Errorf("mru: timestamp contains %d duplicate keys",
			len(ts.ts)-len(keys))
	}

	return nil
}

func (ts *timestamps) Delete(i int) {
	ts.ts = append(ts.ts[:i], ts.ts[i+1:]...)
}

func (ts *timestamps) Dump(w io.Writer) {
	for i := range ts.ts {
		fmt.Fprintf(w, "%d: %s, %d\n", i, ts.K(i), ts.T(i))
	}
}
