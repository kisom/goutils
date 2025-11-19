// Package dbg implements a simple debug printer.
//
// There are two main ways to use it:
//   - By using one of the constructors and calling flag.BoolVar(&debug.Enabled...)
//   - By setting the environment variable GOUTILS_ENABLE_DEBUG to true or false and
//     calling NewFromEnv().
//
// If enabled, any of the print statements will be written to stdout. Otherwise,
// nothing will be emitted.
package dbg

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

const DebugEnvKey = "GOUTILS_ENABLE_DEBUG"

var enabledValues = map[string]bool{
	"1":       true,
	"true":    true,
	"yes":     true,
	"on":      true,
	"y":       true,
	"enable":  true,
	"enabled": true,
}

// A DebugPrinter is a drop-in replacement for fmt.Print*, and also acts as
// an io.WriteCloser when enabled.
type DebugPrinter struct {
	// If Enabled is false, the print statements won't do anything.
	Enabled bool
	out     io.WriteCloser
}

// New returns a new DebugPrinter on os.Stdout.
func New() *DebugPrinter {
	return &DebugPrinter{
		out: os.Stderr,
	}
}

// NewFromEnv returns a new DebugPrinter based on the value of the environment
// variable GOUTILS_ENABLE_DEBUG.
func NewFromEnv() *DebugPrinter {
	enabled := strings.ToLower(os.Getenv(DebugEnvKey))
	return &DebugPrinter{
		out:     os.Stderr,
		Enabled: enabledValues[enabled],
	}
}

// Close satisfies the Closer interface.
func (dbg *DebugPrinter) Close() error {
	return dbg.out.Close()
}

// Write satisfies the Writer interface.
func (dbg *DebugPrinter) Write(p []byte) (int, error) {
	if dbg.Enabled {
		return dbg.out.Write(p)
	}
	return 0, nil
}

// ToFile sets up a new DebugPrinter to a file, truncating it if it exists.
func ToFile(path string) (*DebugPrinter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &DebugPrinter{
		out: file,
	}, nil
}

// To will set up a new DebugPrint to an io.WriteCloser.
func To(w io.WriteCloser) *DebugPrinter {
	return &DebugPrinter{
		out: w,
	}
}

// Print calls fmt.Print if Enabled is true.
func (dbg *DebugPrinter) Print(v ...any) {
	if dbg.Enabled {
		fmt.Fprint(dbg.out, v...)
	}
}

// Println calls fmt.Println if Enabled is true.
func (dbg *DebugPrinter) Println(v ...any) {
	if dbg.Enabled {
		fmt.Fprintln(dbg.out, v...)
	}
}

// Printf calls fmt.Printf if Enabled is true.
func (dbg *DebugPrinter) Printf(format string, v ...any) {
	if dbg.Enabled {
		fmt.Fprintf(dbg.out, format, v...)
	}
}

func (dbg *DebugPrinter) StackTrace() {
	dbg.Write(debug.Stack())
}
