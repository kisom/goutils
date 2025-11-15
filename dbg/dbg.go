// Package dbg implements a debug printer.
package dbg

import (
	"fmt"
	"io"
	"os"
)

// A DebugPrinter is a drop-in replacement for fmt.Print*, and also acts as
// an io.WriteCloser when enabled.
type DebugPrinter struct {
	// If Enabled is false, the print statements won't do anything.
	Enabled bool
	out     io.WriteCloser
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

// New returns a new DebugPrinter on os.Stdout.
func New() *DebugPrinter {
	return &DebugPrinter{
		out: os.Stdout,
	}
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
func (dbg *DebugPrinter) Print(v any) {
	if dbg.Enabled {
		fmt.Fprint(dbg.out, v)
	}
}

// Println calls fmt.Println if Enabled is true.
func (dbg *DebugPrinter) Println(v any) {
	if dbg.Enabled {
		fmt.Fprintln(dbg.out, v)
	}
}

// Printf calls fmt.Printf if Enabled is true.
func (dbg *DebugPrinter) Printf(format string, v any) {
	if dbg.Enabled {
		fmt.Fprintf(dbg.out, format, v)
	}
}
