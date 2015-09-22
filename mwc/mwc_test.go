package mwc

import (
	"bytes"
	"testing"

	"github.com/kisom/testio"
)

func TestMWC(t *testing.T) {
	buf1 := testio.NewBufCloser(nil)
	buf2 := testio.NewBufCloser(nil)

	mwc := MultiWriteCloser(buf1, buf2)

	if _, err := mwc.Write([]byte("hello, world")); err != nil {
		t.Fatalf("%v", err)
	}

	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Fatal("write failed")
	}

	if !bytes.Equal(buf1.Bytes(), []byte("hello, world")) {
		t.Fatal("writing failed")
	}
}
