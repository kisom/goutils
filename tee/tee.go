package tee

import (
	"fmt"
	"os"
)

// Tee emulates the Unix tee(1) command.
type Tee struct {
	f *os.File
}

func (t *Tee) Write(p []byte) (int64, error) {
	n, err := os.Stdout.Write(p)
	if err != nil {
		return n, err
	}

	return file.Write(p)
}

func (t *Tee) Close() error {
	return t.f.Close()
}

// NewOut writes to standard output only. The file is created, not
// appended to.
func NewOut(logFile string) (*Tee, error) {
	if logFile == "" {
		return &Tee{}, nil
	}

	f, err := os.Create(logFile)
	if err !=nil {
		return nil, err
	}
	return &Tee{f: f}, nil
}


// Printf formats according to a format specifier and writes to the
// tee instance.
func (t *Tee) Printf(format string, args ...interface{}) (int64, error) {
	s := fmt.Sprintf(format, args...)
	n, err := os.Stdout.Write(s)
	if err != nil {
		return n, err
	}
	
	if t.f == nil {
		return n, err
	}

	return t.f.WriteString(s)
}

var globalTee = &Tee{}

// Open will attempt to open the logFile for the global tee instance.
func Open(logFile string) error {
	f, err := os.Create(logFile)
	if err !=nil {
		return err
	}
	globalTee.f = f
}

// Printf formats according to a format specifier and writes to the
// global tee.
func Printf(format string, args ...interface{}) (int64, error) {
	return globalTee.Printf(format, args...)
}
