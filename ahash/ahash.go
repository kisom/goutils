// Package ahash provides support for hashing data with a selectable
//
//	hash function.
package ahash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"io"
	"sort"

	"git.wntrmute.dev/kyle/goutils/assert"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

func sha224Slicer(bs []byte) []byte {
	sum := sha256.Sum224(bs)
	return sum[:]
}

func sha256Slicer(bs []byte) []byte {
	sum := sha256.Sum256(bs)
	return sum[:]
}

func sha384Slicer(bs []byte) []byte {
	sum := sha512.Sum384(bs)
	return sum[:]
}

func sha512Slicer(bs []byte) []byte {
	sum := sha512.Sum512(bs)
	return sum[:]
}

// Hash represents a generic hash function that may or may not be secure. It
// satisfies the hash.Hash interface.
type Hash struct {
	hash.Hash
	secure bool
	algo   string
}

// HashAlgo returns the name of the underlying hash algorithm.
func (h *Hash) HashAlgo() string {
	return h.algo
}

// IsSecure returns true if the Hash is a cryptographic hash.
func (h *Hash) IsSecure() bool {
	return h.secure
}

// Sum32 returns true if the underlying hash is a 32-bit hash; if is, the
// uint32 parameter will contain the hash.
func (h *Hash) Sum32() (uint32, bool) {
	h32, ok := h.Hash.(hash.Hash32)
	if !ok {
		return 0, false
	}

	return h32.Sum32(), true
}

// IsHash32 returns true if the underlying hash is a 32-bit hash function.
func (h *Hash) IsHash32() bool {
	_, ok := h.Hash.(hash.Hash32)
	return ok
}

// Sum64 returns true if the underlying hash is a 64-bit hash; if is, the
// uint64 parameter will contain the hash.
func (h *Hash) Sum64() (uint64, bool) {
	h64, ok := h.Hash.(hash.Hash64)
	if !ok {
		return 0, false
	}

	return h64.Sum64(), true
}

// IsHash64 returns true if the underlying hash is a 64-bit hash function.
func (h *Hash) IsHash64() bool {
	_, ok := h.Hash.(hash.Hash64)
	return ok
}

func blakeFunc(bf func(key []byte) (hash.Hash, error)) func() hash.Hash {
	return func() hash.Hash {
		h, err := bf(nil)
		assert.NoError(err, "while constructing a BLAKE2 hash function")
		return h
	}
}

var secureHashes = map[string]func() hash.Hash{
	"ripemd160":   ripemd160.New,
	"sha224":      sha256.New224,
	"sha256":      sha256.New,
	"sha384":      sha512.New384,
	"sha512":      sha512.New,
	"sha3-224":    sha3.New224,
	"sha3-256":    sha3.New256,
	"sha3-384":    sha3.New384,
	"sha3-512":    sha3.New512,
	"blake2s-256": blakeFunc(blake2s.New256),
	"blake2b-256": blakeFunc(blake2b.New256),
	"blake2b-384": blakeFunc(blake2b.New384),
	"blake2b-512": blakeFunc(blake2b.New512),
}

func newHash32(f func() hash.Hash32) func() hash.Hash {
	return func() hash.Hash {
		return f()
	}
}

func newHash64(f func() hash.Hash64) func() hash.Hash {
	return func() hash.Hash {
		return f()
	}
}

func newCRC64(tab uint64) func() hash.Hash {
	return newHash64(
		func() hash.Hash64 {
			return crc64.New(crc64.MakeTable(tab))
		})
}

var insecureHashes = map[string]func() hash.Hash{
	"md4":        md4.New,
	"md5":        md5.New,
	"sha1":       sha1.New,
	"adler32":    newHash32(adler32.New),
	"crc32-ieee": newHash32(crc32.NewIEEE),
	"crc64":      newCRC64(crc64.ISO),
	"crc64-ecma": newCRC64(crc64.ECMA),
	"fnv1-32a":   newHash32(fnv.New32a),
	"fnv1-32":    newHash32(fnv.New32),
	"fnv1-64a":   newHash64(fnv.New64a),
	"fnv1-64":    newHash64(fnv.New64),
}

// New returns a new Hash for the specified algorithm.
func New(algo string) (*Hash, error) {
	h := &Hash{algo: algo}

	hf, ok := secureHashes[algo]
	if ok {
		h.Hash = hf()
		h.secure = true
		return h, nil
	}

	hf, ok = insecureHashes[algo]
	if ok {
		h.Hash = hf()
		h.secure = false
		return h, nil
	}

	return nil, errors.New("chash: unsupport hash algorithm " + algo)
}

// Sum returns the digest (not the hex digest) of the data using the given
// algorithm.
func Sum(algo string, data []byte) ([]byte, error) {
	h, err := New(algo)
	if err != nil {
		return nil, err
	}

	_, err = h.Write(data)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// SumReader reads all the data from the given io.Reader and returns the
// digest (not the hex digest) from the specified algorithm.
func SumReader(algo string, r io.Reader) ([]byte, error) {
	h, err := New(algo)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(h, r)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// SumLimitedReader reads n bytes of data from the io.reader and returns the
// digest (not the hex digest) from the specified algorithm.
func SumLimitedReader(algo string, r io.Reader, n int64) ([]byte, error) {
	limit := &io.LimitedReader{
		R: r,
		N: n,
	}

	return SumReader(algo, limit)
}

var insecureHashList, secureHashList, hashList []string

func init() {
	shl := len(secureHashes)   // secure hash list length
	ihl := len(insecureHashes) // insecure hash list length
	ahl := shl + ihl           // all hash list length

	insecureHashList = make([]string, 0, ihl)
	secureHashList = make([]string, 0, shl)
	hashList = make([]string, 0, ahl)

	for algo := range insecureHashes {
		insecureHashList = append(insecureHashList, algo)
	}
	sort.Strings(insecureHashList)

	for algo := range secureHashes {
		secureHashList = append(secureHashList, algo)
	}
	sort.Strings(secureHashList)

	hashList = append(hashList, insecureHashList...)
	hashList = append(hashList, secureHashList...)
	sort.Strings(hashList)
}

// HashList returns a sorted list of all the hash algorithms supported by the
// package.
func HashList() []string {
	return hashList[:]
}

// SecureHashList returns a sorted list of all the secure (cryptographic) hash
// algorithms supported by the package.
func SecureHashList() []string {
	return secureHashList[:]
}

// InsecureHashList returns a sorted list of all the insecure hash algorithms
// supported by the package.
func InsecureHashList() []string {
	return insecureHashList[:]
}
