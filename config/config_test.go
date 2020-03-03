package config

import (
	"os"
	"testing"
)

const (
	testFilePath = "testdata/test.env"

	// Keys
	kOrder   = "ORDER"
	kSpecies = "SPECIES"
	kName    = "COMMON_NAME"

	// Env
	eOrder   = "corvus"
	eSpecies = "corvus corax"
	eName    = "northern raven"

	// File
	fOrder   = "stringiformes"
	fSpecies = "strix aluco"
	// Name isn't set in the file to test fall through.
)

func init() {
	os.Setenv(kOrder, eOrder)
	os.Setenv(kSpecies, eSpecies)
	os.Setenv(kName, eName)
}

func TestLoadEnvOnly(t *testing.T) {
	order := Get(kOrder)
	species := Get(kSpecies)
	if order != eOrder {
		t.Errorf("want %s, have %s", eOrder, order)
	}

	if species != eSpecies {
		t.Errorf("want %s, have %s", eSpecies, species)
	}
}

func TestLoadFile(t *testing.T) {
	err := LoadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	order := Get(kOrder)
	species := Get(kSpecies)
	name := Get(kName)

	if order != fOrder {
		t.Errorf("want %s, have %s", fOrder, order)
	}

	if species != fSpecies {
		t.Errorf("want %s, have %s", fSpecies, species)
	}

	if name != eName {
		t.Errorf("want %s, have %s", eName, name)
	}
}
