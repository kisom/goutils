package seekbuf

import (
	"fmt"
	"testing"

	"git.wntrmute.dev/kyle/goutils/assert"
)

func TestSeeking(t *testing.T) {
	partA := []byte("hello, ")
	partB := []byte("world!")

	buf := New(partA)
	assert.BoolT(t, buf.Len() == len(partA), fmt.Sprintf("on init: have length %d, want length %d", buf.Len(), len(partA)))

	b := make([]byte, 32)

	n, err := buf.Read(b)
	assert.NoErrorT(t, err)
	assert.BoolT(t, buf.Len() == 0, fmt.Sprintf("after reading 1: have length %d, want length 0", buf.Len()))
	assert.BoolT(t, n == len(partA), fmt.Sprintf("after reading 2: have length %d, want length %d", n, len(partA)))

	n, err = buf.Write(partB)
	assert.NoErrorT(t, err)
	assert.BoolT(t, n == len(partB), fmt.Sprintf("after writing: have length %d, want length %d", n, len(partB)))

	n, err = buf.Read(b)
	assert.NoErrorT(t, err)
	assert.BoolT(t, buf.Len() == 0, fmt.Sprintf("after rereading 1: have length %d, want length 0", buf.Len()))
	assert.BoolT(t, n == len(partB), fmt.Sprintf("after rereading 2: have length %d, want length %d", n, len(partB)))

	partsLen := len(partA) + len(partB)
	buf.Rewind()
	assert.BoolT(t, buf.Len() == partsLen, fmt.Sprintf("after rewinding: have length %d, want length %d", buf.Len(), partsLen))

	buf.Close()
	assert.BoolT(t, buf.Len() == 0, fmt.Sprintf("after closing, have length %d, want length 0", buf.Len()))
}
