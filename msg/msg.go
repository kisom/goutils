// Package msg is a tool for handling commandline output based on
// flags for quiet, verbose, and debug modes. The default is to
// have all modes disabled.
//
// The Qprint messages will only output messages if quiet mode is
// disabled
// The Vprint messages will only output messages if verbose mode
// is enabled.
// The Dprint messages will only output messages if debug mode
// is enabled.
package msg

import (
	"fmt"
	"io"

	"git.wntrmute.dev/kyle/goutils/lib"

	"git.wntrmute.dev/kyle/goutils/dbg"
)

var (
	enableQuiet   bool
	enableVerbose bool
	debug         = dbg.New()
	w             io.Writer
)

func SetQuiet(q bool) {
	enableQuiet = q
}

func SetVerbose(v bool) {
	enableVerbose = v
}

func SetDebug(d bool) {
	debug.Enabled = d
}

func Set(q, v, d bool) {
	SetQuiet(q)
	SetVerbose(v)
	SetDebug(d)
}

func Qprint(a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprint(w, a...)
}

func Qprintf(format string, a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprintf(w, format, a...)
}

func Qprintln(a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprintln(w, a...)
}

func Dprint(a ...any) {
	debug.Print(a...)
}

func Dprintf(format string, a ...any) {
	debug.Printf(format, a...)
}

func Dprintln(a ...any) {
	debug.Println(a...)
}

func StackTrace() {
	debug.StackTrace()
}

func Vprint(a ...any) {
	if !enableVerbose {
		return
	}

	fmt.Fprint(w, a...)
}

func Vprintf(format string, a ...any) {
	if !enableVerbose {
		return
	}

	fmt.Fprintf(w, format, a...)
}

func Vprintln(a ...any) {
	if !enableVerbose {
		return
	}

	fmt.Fprintln(w, a...)
}

func Print(a ...any) {
	fmt.Fprint(w, a...)
}

func Printf(format string, a ...any) {
	fmt.Fprintf(w, format, a...)
}

func Println(a ...any) {
	fmt.Fprintln(w, a...)
}

// SetWriter changes the output for messages.
func SetWriter(dst io.Writer) {
	w = dst
	dbgEnabled := debug.Enabled
	debug = dbg.To(lib.WithCloser(w))
	debug.Enabled = dbgEnabled
}
