package tee

import (
	"fmt"
	"os"
)

type WriteStringCloser interface {
	Write([]byte) (int, error)
	WriteString(string) (int, error)
	Close() error
}

// Tee emulates the Unix tee(1) command.
type Tee struct {
	f       WriteStringCloser
	Verbose bool
}

// NewOut writes to standard output only. The file is created, not
// appended to.
func NewOut(logFile string) (*Tee, error) {
	if logFile == "" {
		return &Tee{}, nil
	}

	f, err := os.Create(logFile)
	if err != nil {
		return nil, err
	}
	return &Tee{f: f}, nil
}

func (t *Tee) Write(p []byte) (int, error) {
	n, err := os.Stdout.Write(p)
	if err != nil {
		return n, err
	}

	if t.f != nil {
		return t.f.Write(p)
	}
	return n, nil
}

// Close calls Close on the underlying file if present.
// It is safe to call Close on a Tee with no file; in that case, it returns nil.
func (t *Tee) Close() error {
	if t == nil || t.f == nil {
		return nil
	}
	err := t.f.Close()
	t.f = nil
	return err
}

// Printf formats according to a format specifier and writes to the
// tee instance.
func (t *Tee) Printf(format string, args ...any) (int, error) {
	s := fmt.Sprintf(format, args...)
	n, err := os.Stdout.WriteString(s)
	if err != nil {
		return n, err
	}

	if t.f == nil {
		return n, err
	}

	return t.f.WriteString(s)
}

// VPrintf is a variant of Printf that only prints if the Tee's
// Verbose flag is set.
func (t *Tee) VPrintf(format string, args ...any) (int, error) {
	if t.Verbose {
		return t.Printf(format, args...)
	}
	return 0, nil
}

var globalTee = &Tee{}

// Open will attempt to open the logFile for the global tee instance.
func Open(logFile string) error {
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	globalTee.f = f
	return nil
}

// Printf formats according to a format specifier and writes to the
// global tee.
func Printf(format string, args ...any) (int, error) {
	return globalTee.Printf(format, args...)
}

// VPrintf calls VPrintf on the global tee instance.
func VPrintf(format string, args ...any) (int, error) {
	return globalTee.VPrintf(format, args...)
}

// Close calls close on the global tee instance.
func Close() error {
	return globalTee.Close()
}

// SetVerbose controls the verbosity of the global tee.
func SetVerbose(verbose bool) {
	globalTee.Verbose = verbose
}
