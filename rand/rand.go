// Package rand contains utilities for interacting with math/rand, including
// seeding from a random sed.
package rand

import (
	"crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

// CryptoUint64 generates a cryptographically-secure 64-bit integer.
func CryptoUint64() (uint64, error) {
	bs := make([]byte, 8)
	_, err := rand.Read(bs)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(bs), nil
}

// Seed initialises the non-cryptographic PRNG with a random,
// cryptographically secure value. This is done just as a good
// way to make this random. The returned 64-bit value is the seed.
func Seed() (uint64, error) {
	seed, err := CryptoUint64()
	if err != nil {
		return 0, err
	}

	// NB: this is permitted.
	mrand.Seed(int64(seed))
	return seed, nil
}

// Int is a wrapper for math.Int so only one package needs to be imported.
func Int() int {
	return mrand.Int()
}

// Intn is a wrapper for math.Intn so only one package needs to be imported.
func Intn(max int) int {
	return mrand.Intn(max)
}

// Intn2 returns a random value between min and max, inclusive.
func Intn2(min, max int) int {
	return Intn(max-min) + min
}
