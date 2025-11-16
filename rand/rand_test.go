package rand

import (
	"fmt"
	mrand "math/rand"
	"testing"
)

func TestCryptoUint64(t *testing.T) {
	n1, err := CryptoUint64()
	if err != nil {
		t.Fatal(err)
	}

	n2, err := CryptoUint64()
	if err != nil {
		t.Fatal(err)
	}

	// This has such a low chance of occurring that it's likely to be
	// indicative of a bad CSPRNG.
	if n1 == n2 {
		t.Fatalf("repeated random uint64s: %d", n1)
	}
}

func TestIntn(t *testing.T) {
	expected := []int{3081, 4887, 4847, 1059, 3081}
	mrand.Seed(1)
	for i := range 5 {
		n := Intn2(1000, 5000)

		if n != expected[i] {
			fmt.Printf("invalid sequence at %d: expected %d, have %d", i, expected[i], n)
		}
	}
}

func TestSeed(t *testing.T) {
	seed1, err := Seed()
	if err != nil {
		t.Fatal(err)
	}

	var seed2 uint64
	n1 := Int()
	tries := 0

	for {
		seed2, err = Seed()
		if err != nil {
			t.Fatal(err)
		}

		if seed1 != seed2 {
			break
		}

		tries++

		if tries > 3 {
			t.Fatal("can't generate two unique seeds")
		}
	}

	n2 := Int()

	// Again, this not impossible, merely statistically improbably and a
	// potential canary for RNG issues.
	if n1 == n2 {
		t.Fatalf("repeated integers fresh from two unique seeds: %d/%d -> %d",
			seed1, seed2, n1)
	}
}
