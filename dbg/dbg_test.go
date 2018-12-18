package dbg

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kisom/goutils/testio"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	buf := testio.NewBufCloser(nil)
	dbg := New()
	dbg.out = buf

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")
	require.Equal(t, 0, buf.Len())

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11
	require.Equal(t, 22, buf.Len())

	err := dbg.Close()
	require.NoError(t, err)
}

func TestTo(t *testing.T) {
	buf := testio.NewBufCloser(nil)
	dbg := To(buf)

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")
	require.Equal(t, 0, buf.Len())

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11

	require.Equal(t, 22, buf.Len())

	err := dbg.Close()
	require.NoError(t, err)
}

func TestToFile(t *testing.T) {
	testFile, err := ioutil.TempFile("", "dbg")
	require.NoError(t, err)
	err = testFile.Close()
	require.NoError(t, err)

	testFileName := testFile.Name()
	defer os.Remove(testFileName)

	dbg, err := ToFile(testFileName)
	require.NoError(t, err)

	dbg.Print("hello")
	dbg.Println("hello")
	dbg.Printf("hello %s", "world")

	stat, err := os.Stat(testFileName)
	require.NoError(t, err)

	require.EqualValues(t, 0, stat.Size())

	dbg.Enabled = true
	dbg.Print("hello")              // +5
	dbg.Println("hello")            // +6
	dbg.Printf("hello %s", "world") // +11

	stat, err = os.Stat(testFileName)
	require.NoError(t, err)

	require.EqualValues(t, 22, stat.Size())

	err = dbg.Close()
	require.NoError(t, err)
}

func TestWriting(t *testing.T) {
	data := []byte("hello, world")
	buf := testio.NewBufCloser(nil)
	dbg := To(buf)

	n, err := dbg.Write(data)
	require.NoError(t, err)
	require.EqualValues(t, 0, n)

	dbg.Enabled = true
	n, err = dbg.Write(data)
	require.NoError(t, err)
	require.EqualValues(t, 12, n)

	err = dbg.Close()
	require.NoError(t, err)
}

func TestToFileError(t *testing.T) {
	testFile, err := ioutil.TempFile("", "dbg")
	require.NoError(t, err)
	err = testFile.Chmod(0400)
	require.NoError(t, err)
	err = testFile.Close()
	require.NoError(t, err)

	testFileName := testFile.Name()

	_, err = ToFile(testFileName)
	require.Error(t, err)

	err = os.Remove(testFileName)
	require.NoError(t, err)
}
