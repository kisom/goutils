package config_test

import (
	"os"
	"testing"

	"git.wntrmute.dev/kyle/goutils/config"
)

const (
	testFilePath = "testdata/test.env"

	// Key constants.
	kOrder   = "ORDER"
	kSpecies = "SPECIES"
	kName    = "COMMON_NAME"

	eOrder   = "corvus"
	eSpecies = "corvus corax"
	eName    = "northern raven"

	fOrder   = "stringiformes"
	fSpecies = "strix aluco"
)

func init() {
	os.Setenv(kOrder, eOrder)
	os.Setenv(kSpecies, eSpecies)
	os.Setenv(kName, eName)
}

func TestLoadEnvOnly(t *testing.T) {
	order := config.Get(kOrder)
	species := config.Get(kSpecies)
	if order != eOrder {
		t.Errorf("want %s, have %s", eOrder, order)
	}

	if species != eSpecies {
		t.Errorf("want %s, have %s", eSpecies, species)
	}
}

func TestLoadFile(t *testing.T) {
	err := config.LoadFile(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	order := config.Get(kOrder)
	species := config.Get(kSpecies)
	name := config.Get(kName)

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
