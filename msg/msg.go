// Package msg is a tool for handling commandline output based on
// flags for quiet, verbose, and debug modes. The default is to
// have all modes disabled.
//
// The QPrint messages will only output messages if quiet mode is
// disabled
// The VPrint messages will only output messages if verbose mode
// is enabled.
// The DPrint messages will only output messages if debug mode
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
	debug         *dbg.DebugPrinter
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

func QPrint(a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprint(w, a...)
}

func QPrintf(format string, a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprintf(w, format, a...)
}

func QPrintln(a ...any) {
	if enableQuiet {
		return
	}

	fmt.Fprintln(w, a...)
}

func DPrint(a ...any) {
	debug.Print(a...)
}

func DPrintf(format string, a ...any) {
	debug.Printf(format, a...)
}

func DPrintln(a ...any) {
	debug.Println(a...)
}

func StackTrace() {
	debug.StackTrace()
}

func VPrint(a ...any) {
	if !enableVerbose {
		return
	}

	fmt.Fprint(w, a...)
}

func VPrintf(format string, a ...any) {
	if !enableVerbose {
		return
	}

	fmt.Fprintf(w, format, a...)
}

func VPrintln(a ...any) {
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
	debug = dbg.To(lib.WithCloser(w))
}
