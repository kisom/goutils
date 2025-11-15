// Package backoff contains an implementation of an intelligent backoff
// strategy. It is based on the approach in the AWS architecture blog
// article titled "Exponential Backoff And Jitter", which is found at
// http://www.awsarchitectureblog.com/2015/03/backoff.html.
//
// Essentially, the backoff has an interval `time.Duration`; the nth
// call to backoff will return a `time.Duration` that is 2^n *
// interval. If jitter is enabled (which is the default behaviour),
// the duration is a random value between 0 and 2^n * interval.  The
// backoff is configured with a maximum duration that will not be
// exceeded.
//
// This package uses math/rand/v2 for jitter, which is automatically
// seeded from a cryptographically secure source.
package backoff

import (
	"math"
	"math/rand/v2"
	"time"
)

// DefaultInterval is used when a Backoff is initialised with a
// zero-value Interval.
var DefaultInterval = 5 * time.Minute

// DefaultMaxDuration is the maximum amount of time that the backoff will
// delay for.
var DefaultMaxDuration = 6 * time.Hour

// A Backoff contains the information needed to intelligently backoff
// and retry operations using an exponential backoff algorithm. It should
// be initialised with a call to `New`.
//
// Only use a Backoff from a single goroutine, it is not safe for concurrent
// access.
type Backoff struct {
	// maxDuration is the largest possible duration that can be
	// returned from a call to Duration.
	maxDuration time.Duration

	// interval controls the time step for backing off.
	interval time.Duration

	// noJitter controls whether to use the "Full Jitter" improvement to attempt
	// to smooth out spikes in a high-contention scenario. If noJitter is set to
	// true, no jitter will be introduced.
	noJitter bool

	// decay controls the decay of n. If it is non-zero, n is
	// reset if more than the last backoff + decay has elapsed since
	// the last try.
	decay time.Duration

	n       uint64
	lastTry time.Time
}

// New creates a new backoff with the specified maxDuration duration and
// interval. Zero values may be used to use the default values.
//
// Panics if either dMax or interval is negative.
func New(dMax time.Duration, interval time.Duration) *Backoff {
	if dMax < 0 || interval < 0 {
		panic("backoff: dMax or interval is negative")
	}

	b := &Backoff{
		maxDuration: dMax,
		interval:    interval,
	}
	b.setup()
	return b
}

// NewWithoutJitter works similarly to New, except that the created
// Backoff will not use jitter.
func NewWithoutJitter(dMax time.Duration, interval time.Duration) *Backoff {
	b := New(dMax, interval)
	b.noJitter = true
	return b
}

func (b *Backoff) setup() {
	if b.interval == 0 {
		b.interval = DefaultInterval
	}

	if b.maxDuration == 0 {
		b.maxDuration = DefaultMaxDuration
	}
}

// Duration returns a time.Duration appropriate for the backoff,
// incrementing the attempt counter.
func (b *Backoff) Duration() time.Duration {
	b.setup()

	b.decayN()

	d := b.duration(b.n)

	if b.n < math.MaxUint64 {
		b.n++
	}

	if !b.noJitter {
		d = time.Duration(rand.Int64N(int64(d))) // #nosec G404
	}

	return d
}

const maxN uint64 = 63

// requires b to be locked.
func (b *Backoff) duration(n uint64) time.Duration {
	// Use left shift on the underlying integer representation to avoid
	// multiplying time.Duration by time.Duration (which is semantically
	// incorrect and flagged by linters).
	if n >= maxN {
		// Saturate when n would overflow a 64-bit shift or exceed maxDuration.
		return b.maxDuration
	}

	// Calculate 2^n * interval using a shift. Detect overflow by checking
	// for sign change or monotonicity loss and clamp to maxDuration.
	shifted := b.interval << n
	if shifted < 0 || shifted < b.interval {
		// Overflow occurred during the shift; clamp to maxDuration.
		return b.maxDuration
	}

	if shifted > b.maxDuration {
		return b.maxDuration
	}

	return shifted
}

// Reset resets the attempt counter of a backoff.
//
// It should be called when the rate-limited action succeeds.
func (b *Backoff) Reset() {
	b.lastTry = time.Time{}
	b.n = 0
}

// SetDecay sets the duration after which the try counter will be reset.
// Panics if decay is smaller than 0.
//
// The decay only kicks in if at least the last backoff + decay has elapsed
// since the last try.
func (b *Backoff) SetDecay(decay time.Duration) {
	if decay < 0 {
		panic("backoff: decay < 0")
	}

	b.decay = decay
}

// requires b to be locked.
func (b *Backoff) decayN() {
	if b.decay == 0 {
		return
	}

	if b.lastTry.IsZero() {
		b.lastTry = time.Now()
		return
	}

	lastDuration := b.duration(b.n - 1)
	// Reset when the elapsed time is at least the previous backoff plus decay.
	// Using ">=" avoids boundary flakiness in tests and real usage.
	decayed := time.Since(b.lastTry) >= lastDuration+b.decay
	b.lastTry = time.Now()

	if !decayed {
		return
	}

	b.n = 0
}
