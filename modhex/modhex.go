// Package modhex implements the modified hexadecimal encoding as used
// by Yubico in their series of products.
package modhex

import "fmt"

// Encoding is a mapping of hexadecimal values to a new byte value.
// This means that the encoding for a single byte is two bytes.
type Encoding struct {
	decoding map[byte]byte
	encoding [16]byte
}

// A CorruptInputError is returned if the input string contains
// invalid characters for the encoding or if the input is the wrong
// length. It contains the number of bytes written out.
type CorruptInputError struct {
	written int64
}

func (err CorruptInputError) Error() string {
	return fmt.Sprintf("modhex: corrupt input at byte %d", err.written)
}

func (err CorruptInputError) Written() int64 {
	return err.written
}

var encodeStd = "cbdefghijklnrtuv"

// NewEncoding builds a new encoder from the alphabet passed in,
// which must be a 16-byte string.
func NewEncoding(encoder string) *Encoding {
	if len(encoder) != 16 {
		return nil
	}

	enc := new(Encoding)
	enc.decoding = make(map[byte]byte)
	for i := range encoder {
		enc.encoding[i] = encoder[i]
		enc.decoding[encoder[i]] = byte(i)
	}
	return enc
}

// StdEncoding is the canonical modhex alphabet as used by Yubico.
var StdEncoding = NewEncoding(encodeStd)

// Encode encodes src to dst, writing at most EncodedLen(len(src))
// bytes to dst.
func (enc *Encoding) Encode(dst, src []byte) {
	out := dst

	for i := 0; i < len(src) && len(out) > 1; i++ {
		var b [2]byte
		b[0] = enc.encoding[(src[i]&0xf0)>>4]
		b[1] = enc.encoding[src[i]&0xf]
		copy(out[:2], b[:])
		out = out[2:]
	}
}

// EncodedLen returns the encoded length of a buffer of n bytes.
func EncodedLen(n int) int {
	return n << 1
}

// DecodedLen returns the decoded length of a buffer of n bytes.
func DecodedLen(n int) int {
	return n >> 1
}

// Decode decodes src into dst, which will be at most DecodedLen(len(src)).
// It returns the number of bytes written, and any error that occurred.
func (enc *Encoding) Decode(dst, src []byte) (n int, err error) {
	out := dst

	for i := 0; i < len(src); i += 2 {
		if (len(src) - i) < 2 {
			return i >> 1, CorruptInputError{int64(i >> 1)}
		}
		var b byte
		if high, ok := enc.decoding[src[i]]; !ok {
			return i >> 1, CorruptInputError{int64(i >> 1)}
		} else if low, ok := enc.decoding[src[i+1]]; !ok {
			return i >> 1, CorruptInputError{int64(i >> 1)}
		} else {
			b = high << 4
			b += low
			out[0] = b
			out = out[1:]
		}
	}
	return len(dst), nil
}

// EncodeToString is a convenience function to encode src as a
// string.
func (enc *Encoding) EncodeToString(src []byte) string {
	dst := make([]byte, EncodedLen(len(src)))
	enc.Encode(dst, src)
	return string(dst)
}

// DecodeString decodes the string passed in as its decoded bytes.
func (enc *Encoding) DecodeString(s string) ([]byte, error) {
	dst := make([]byte, DecodedLen(len(s)))
	src := []byte(s)
	_, err := enc.Decode(dst, src)
	return dst, err
}
