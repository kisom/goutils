package ahash

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/kisom/goutils/assert"
)

func TestSecureHash(t *testing.T) {
	algo := "sha256"
	h, err := New(algo)
	assert.NoErrorT(t, err)
	assert.BoolT(t, h.IsSecure(), algo+" should be a secure hash")
	assert.BoolT(t, h.HashAlgo() == algo, "hash returned the wrong HashAlgo")
	assert.BoolT(t, !h.IsHash32(), algo+" isn't actually a 32-bit hash")
	assert.BoolT(t, !h.IsHash64(), algo+" isn't actually a 64-bit hash")

	var data []byte
	var expected = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sum, err := Sum(algo, data)
	assert.NoErrorT(t, err)
	assert.BoolT(t, fmt.Sprintf("%x", sum) == expected, fmt.Sprintf("expected hash %s but have %x", expected, sum))

	data = []byte("hello, world")
	buf := bytes.NewBuffer(data)
	expected = "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b"
	sum, err = SumReader(algo, buf)
	assert.NoErrorT(t, err)
	assert.BoolT(t, fmt.Sprintf("%x", sum) == expected, fmt.Sprintf("expected hash %s but have %x", expected, sum))

	data = []byte("hello world")
	_, err = h.Write(data)
	assert.NoErrorT(t, err)
	unExpected := "09ca7e4eaa6e8ae9c7d261167129184883644d07dfba7cbfbc4c8a2e08360d5b"
	sum = h.Sum(nil)
	assert.BoolT(t, fmt.Sprintf("%x", sum) != unExpected, fmt.Sprintf("hash shouldn't have returned %x", unExpected))
}

func TestInsecureHash(t *testing.T) {
	algo := "md5"
	h, err := New(algo)
	assert.NoErrorT(t, err)
	assert.BoolT(t, !h.IsSecure(), algo+" shouldn't be a secure hash")
	assert.BoolT(t, h.HashAlgo() == algo, "hash returned the wrong HashAlgo")
	assert.BoolT(t, !h.IsHash32(), algo+" isn't actually a 32-bit hash")
	assert.BoolT(t, !h.IsHash64(), algo+" isn't actually a 64-bit hash")

	var data []byte
	var expected = "d41d8cd98f00b204e9800998ecf8427e"
	sum, err := Sum(algo, data)
	assert.NoErrorT(t, err)
	assert.BoolT(t, fmt.Sprintf("%x", sum) == expected, fmt.Sprintf("expected hash %s but have %x", expected, sum))

	data = []byte("hello, world")
	buf := bytes.NewBuffer(data)
	expected = "e4d7f1b4ed2e42d15898f4b27b019da4"
	sum, err = SumReader(algo, buf)
	assert.NoErrorT(t, err)
	assert.BoolT(t, fmt.Sprintf("%x", sum) == expected, fmt.Sprintf("expected hash %s but have %x", expected, sum))

	data = []byte("hello world")
	_, err = h.Write(data)
	assert.NoErrorT(t, err)
	unExpected := "e4d7f1b4ed2e42d15898f4b27b019da4"
	sum = h.Sum(nil)
	assert.BoolT(t, fmt.Sprintf("%x", sum) != unExpected, fmt.Sprintf("hash shouldn't have returned %x", unExpected))
}

func TestHash32(t *testing.T) {
	algo := "crc32-ieee"
	h, err := New(algo)
	assert.NoErrorT(t, err)
	assert.BoolT(t, !h.IsSecure(), algo+" shouldn't be a secure hash")
	assert.BoolT(t, h.HashAlgo() == algo, "hash returned the wrong HashAlgo")
	assert.BoolT(t, h.IsHash32(), algo+" is actually a 32-bit hash")
	assert.BoolT(t, !h.IsHash64(), algo+" isn't actually a 64-bit hash")

	var data []byte
	var expected uint32

	h.Write(data)
	sum, ok := h.Sum32()
	assert.BoolT(t, ok, algo+" should be able to return a Sum32")
	assert.BoolT(t, expected == sum, fmt.Sprintf("%s returned the %d but expected %d", algo, sum, expected))

	data = []byte("hello, world")
	expected = 0xffab723a
	h.Write(data)
	sum, ok = h.Sum32()
	assert.BoolT(t, ok, algo+" should be able to return a Sum32")
	assert.BoolT(t, expected == sum, fmt.Sprintf("%s returned the %d but expected %d", algo, sum, expected))

	h.Reset()
	data = []byte("hello world")
	h.Write(data)
	sum, ok = h.Sum32()
	assert.BoolT(t, ok, algo+" should be able to return a Sum32")
	assert.BoolT(t, expected != sum, fmt.Sprintf("%s returned %d but shouldn't have", algo, sum))
}

func TestHash64(t *testing.T) {
	algo := "crc64"
	h, err := New(algo)
	assert.NoErrorT(t, err)
	assert.BoolT(t, !h.IsSecure(), algo+" shouldn't be a secure hash")
	assert.BoolT(t, h.HashAlgo() == algo, "hash returned the wrong HashAlgo")
	assert.BoolT(t, h.IsHash64(), algo+" is actually a 64-bit hash")
	assert.BoolT(t, !h.IsHash32(), algo+" isn't actually a 32-bit hash")

	var data []byte
	var expected uint64

	h.Write(data)
	sum, ok := h.Sum64()
	assert.BoolT(t, ok, algo+" should be able to return a Sum64")
	assert.BoolT(t, expected == sum, fmt.Sprintf("%s returned the %d but expected %d", algo, sum, expected))

	data = []byte("hello, world")
	expected = 0x16c45c0eb1d9c2ec
	h.Write(data)
	sum, ok = h.Sum64()
	assert.BoolT(t, ok, algo+" should be able to return a Sum64")
	assert.BoolT(t, expected == sum, fmt.Sprintf("%s returned the %d but expected %d", algo, sum, expected))

	h.Reset()
	data = []byte("hello world")
	h.Write(data)
	sum, ok = h.Sum64()
	assert.BoolT(t, ok, algo+" should be able to return a Sum64")
	assert.BoolT(t, expected != sum, fmt.Sprintf("%s returned %d but shouldn't have", algo, sum))
}

func TestListLengthSanity(t *testing.T) {
	all := HashList()
	secure := SecureHashList()
	insecure := InsecureHashList()

	assert.BoolT(t, len(all) == len(secure)+len(insecure))
}
