package iniconf_test

import (
	"errors"
	"os"
	"sort"
	"testing"

	"git.wntrmute.dev/kyle/goutils/config/iniconf"
)

// FailWithError is a utility for dumping errors and failing the test.
func FailWithError(t *testing.T, err error) {
	t.Log("failed")
	if err != nil {
		t.Log("[!] ", err.Error())
	}
	t.FailNow()
}

// UnlinkIfExists removes a file if it exists.
func UnlinkIfExists(file string) {
	_, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		panic("failed to remove " + file)
	}
	os.Remove(file)
}

// stringSlicesEqual compares two string lists, checking that they
// contain the same elements.
func stringSlicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	for i := range slice2 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func TestGoodConfig(t *testing.T) {
	testFile := "testdata/test.conf"
	t.Logf("[+] validating known-good config... ")
	cmap, err := iniconf.ParseFile(testFile)
	if err != nil {
		FailWithError(t, err)
	} else if len(cmap) != 2 {
		FailWithError(t, err)
	}
	t.Log("ok")
}

func TestGoodConfig2(t *testing.T) {
	testFile := "testdata/test2.conf"
	t.Logf("[+] validating second known-good config... ")
	cmap, err := iniconf.ParseFile(testFile)
	switch {
	case err != nil:
		FailWithError(t, err)
	case len(cmap) != 1:
		FailWithError(t, err)
	case len(cmap["default"]) != 3:
		FailWithError(t, err)
	default:
		// nothing to do here
	}
	t.Log("ok")
}

func TestBadConfig(t *testing.T) {
	testFile := "testdata/bad.conf"
	t.Logf("[+] ensure invalid config file fails... ")
	_, err := iniconf.ParseFile(testFile)
	if err == nil {
		err = errors.New("invalid config file should fail")
		FailWithError(t, err)
	}
	t.Log("ok")
}

func TestWriteConfigFile(t *testing.T) {
	t.Logf("[+] ensure config file is written properly... ")
	const testFile = "testdata/test.conf"
	const testOut = "testdata/test.out"

	cmap, err := iniconf.ParseFile(testFile)
	if err != nil {
		FailWithError(t, err)
	}

	defer UnlinkIfExists(testOut)
	err = cmap.WriteFile(testOut)
	if err != nil {
		FailWithError(t, err)
	}

	cmap2, err := iniconf.ParseFile(testOut)
	if err != nil {
		FailWithError(t, err)
	}

	sectionList1 := cmap.ListSections()
	sectionList2 := cmap2.ListSections()
	sort.Strings(sectionList1)
	sort.Strings(sectionList2)
	if !stringSlicesEqual(sectionList1, sectionList2) {
		err = errors.New("section lists don't match")
		FailWithError(t, err)
	}

	for _, section := range sectionList1 {
		for _, k := range cmap[section] {
			if cmap[section][k] != cmap2[section][k] {
				err = errors.New("config key doesn't match")
				FailWithError(t, err)
			}
		}
	}
	t.Log("ok")
}

func TestQuotedValue(t *testing.T) {
	testFile := "testdata/test.conf"
	t.Logf("[+] validating quoted value... ")
	cmap, _ := iniconf.ParseFile(testFile)
	val := cmap["sectionName"]["key4"]
	if val != " space at beginning and end " {
		FailWithError(t, errors.New("Wrong value in double quotes ["+val+"]"))
	}

	val = cmap["sectionName"]["key5"]
	if val != " is quoted with single quotes " {
		FailWithError(t, errors.New("Wrong value in single quotes ["+val+"]"))
	}
	t.Log("ok")
}
