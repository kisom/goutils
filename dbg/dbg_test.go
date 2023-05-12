package dbg

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"git.wntrmute.dev/kyle/goutils/assert"
	"git.wntrmute.dev/kyle/goutils/testio"
)

func TestNew(t *testing.T) {
	buf := testio.NewBufCloser(nil)
	dbg := New()
	dbg.out = buf

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")
	assert.BoolT(t, buf.Len() == 0)

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11
	assert.BoolT(t, buf.Len() == 22, fmt.Sprintf("buffer should be length 22 but is length %d", buf.Len()))

	err := dbg.Close()
	assert.NoErrorT(t, err)
}

func TestTo(t *testing.T) {
	buf := testio.NewBufCloser(nil)
	dbg := To(buf)

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")
	assert.BoolT(t, buf.Len() == 0, "debug output should be suppressed")

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11
	assert.BoolT(t, buf.Len() == 22, "didn't get the expected debug output")

	err := dbg.Close()
	assert.NoErrorT(t, err)
}

func TestToFile(t *testing.T) {
	testFile, err := ioutil.TempFile("", "dbg")
	assert.NoErrorT(t, err)
	err = testFile.Close()
	assert.NoErrorT(t, err)

	testFileName := testFile.Name()
	defer os.Remove(testFileName)

	dbg, err := ToFile(testFileName)
	assert.NoErrorT(t, err)

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")

	stat, err := os.Stat(testFileName)
	assert.NoErrorT(t, err)

	assert.BoolT(t, stat.Size() == 0, "no debug output should have been sent to the log file")

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11

	stat, err = os.Stat(testFileName)
	assert.NoErrorT(t, err)

	assert.BoolT(t, stat.Size() == 22, fmt.Sprintf("have %d bytes in the log file, expected 22", stat.Size()))

	err = dbg.Close()
	assert.NoErrorT(t, err)
}

func TestWriting(t *testing.T) {
	data := []byte("hello, world")
	buf := testio.NewBufCloser(nil)
	dbg := To(buf)

	n, err := dbg.Write(data)
	assert.NoErrorT(t, err)
	assert.BoolT(t, n == 0, "expected nothing to be written to the buffer")

	dbg.Enabled = true
	n, err = dbg.Write(data)
	assert.NoErrorT(t, err)
	assert.BoolT(t, n == 12, fmt.Sprintf("wrote %d bytes in the buffer, expected to write 12", n))

	err = dbg.Close()
	assert.NoErrorT(t, err)
}

func TestToFileError(t *testing.T) {
	testFile, err := ioutil.TempFile("", "dbg")
	assert.NoErrorT(t, err)
	err = testFile.Chmod(0400)
	assert.NoErrorT(t, err)
	err = testFile.Close()
	assert.NoErrorT(t, err)

	testFileName := testFile.Name()

	_, err = ToFile(testFileName)
	assert.ErrorT(t, err)

	err = os.Remove(testFileName)
	assert.NoErrorT(t, err)
}
