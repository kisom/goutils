package sbuf_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"git.wntrmute.dev/kyle/goutils/sbuf"
	"golang.org/x/crypto/nacl/box"
)

var (
	buf          = &sbuf.Buffer{}
	testMessage1 = []byte("round and round and round we go, where we stop, no one knows")
	testMessage2 = []byte("the deconstruction of falling stars")
)

func TestWrite(t *testing.T) {
	n, err := buf.Write(testMessage1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(testMessage1) {
		t.Fatalf("expected to write %d bytes, but only wrote %d bytes", len(testMessage1), n)
	}

	if buf.Len() != len(testMessage1) {
		t.Fatalf("expected a length of %d, but have a length of %d", len(testMessage1), buf.Len())
	}

	if buf.Cap() != (len(testMessage1) * 2) {
		t.Fatalf("expected a capacity of %d, but have a capacity of %d", len(testMessage1)*2, buf.Cap())
	}

	n, err = buf.Write(testMessage2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(testMessage2) {
		t.Fatalf("expected to write %d bytes, but only wrote %d bytes", len(testMessage2), n)
	}

	n, err = buf.Write(testMessage2)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(testMessage2) {
		t.Fatalf("expected to write %d bytes, but only wrote %d bytes", len(testMessage2), n)
	}
}

func TestRead(t *testing.T) {
	var p = make([]byte, len(testMessage1))
	n, err := buf.Read(p)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(testMessage1) {
		t.Fatalf("expected to read %d bytes, but only read %d bytes", len(testMessage1), n)
	}

	if !bytes.Equal(p, testMessage1) {
		t.Fatalf("expected p='%s', but p='%s'", testMessage1, p)
	}

	p = make([]byte, len(testMessage2))
	n, err = buf.Read(p)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(testMessage2) {
		t.Fatalf("expected to read %d bytes, but only read %d bytes", len(testMessage2), n)
	}

	if !bytes.Equal(p, testMessage2) {
		t.Fatalf("expected p='%s', but p='%s'", testMessage2, p)
	}

	buf.Close()
	n, err = buf.Read(p)
	if err == nil {
		t.Fatal("expected EOF")
	} else if n != 0 {
		t.Fatalf("expect n=0, but n=%d", n)
	}

	p = nil
	n, err = buf.Read(p)
	if err != nil {
		t.Fatalf("%v", err)
	} else if n != 0 {
		t.Fatalf("expect n=0, but n=%d", n)
	}
}

func TestShortRead(t *testing.T) {
	tmpMessage := []byte("hello, world")
	buf.Write(tmpMessage)

	var p = make([]byte, len(testMessage1))
	n, err := buf.Read(p)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if n != len(tmpMessage) {
		t.Fatalf("expected to read %d bytes, but only read %d bytes", len(testMessage1), n)
	}
}

func TestNewBuffer(t *testing.T) {
	testBuffer := sbuf.NewBuffer(32)
	if testBuffer.Len() != 0 {
		t.Fatalf("expected new buffer length to be 0, have %d",
			testBuffer.Len())
	}

	if testBuffer.Cap() != 32 {
		t.Fatalf("expected new buffer capacity to be 0, have %d",
			testBuffer.Cap())
	}
}

func TestNewBufferFrom(t *testing.T) {
	p := make([]byte, len(testMessage1))
	copy(p, testMessage1)
	testBuffer := sbuf.NewBufferFrom(p)
	if !bytes.Equal(testBuffer.Bytes(), testMessage1) {
		t.Fatal("new buffer wasn't constructed properly")
	}
}

func TestBytes(t *testing.T) {
	p := make([]byte, len(testMessage1))
	copy(p, testMessage1)
	testBuffer := sbuf.NewBufferFrom(p)

	out := testBuffer.Bytes()
	if testBuffer.Len() != 0 {
		t.Fatal("buffer was not closed")
	}

	if !bytes.Equal(out, testMessage1) {
		t.Fatal("buffer did not return the right data")
	}

	out = testBuffer.Bytes()
	if out != nil {
		t.Fatal("a closed buffer should return nil for Bytes")
	}
}

func TestRWByte(t *testing.T) {
	testBuffer := sbuf.NewBuffer(0)
	c := byte(42)
	err := testBuffer.WriteByte(c)
	if err != nil {
		t.Fatalf("%v", err)
	}

	c, err = testBuffer.ReadByte()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if c != 42 {
		t.Fatalf("Expected 42, have %d", c)
	}

	_, err = testBuffer.ReadByte()
	if err == nil {
		t.Fatal("Expected EOF")
	}
}

func BenchmarkRead(b *testing.B) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("%v", err)
	}
	b.ReportAllocs()

	for b.Loop() {
		_, err = buf.Write(priv[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		_, err = buf.Write(pub[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		b.SetBytes(64)
	}
}

func BenchmarkFixed(b *testing.B) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("%v", err)
	}

	buf = sbuf.NewBuffer(64 * b.N)
	b.ReportAllocs()

	for b.Loop() {
		_, err = buf.Write(priv[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		_, err = buf.Write(pub[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		b.SetBytes(64)
	}
}

func BenchmarkWrite(b *testing.B) {
	var pub = new([32]byte)
	var priv = new([32]byte)

	b.ReportAllocs()

	for i := 0; i < b.N && buf.Len() >= 64; i++ {
		_, err := buf.Read(priv[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		_, err = buf.Read(pub[:])
		if err != nil {
			b.Fatalf("%v", err)
		}
		b.SetBytes(64)
	}
}
