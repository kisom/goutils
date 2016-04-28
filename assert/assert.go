// Package assert provides C-like assertions for Go. For more
// information, see assert(3) (e.g. `man 3 assert`).
package assert

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

// NoDebug, if set to true, will cause all asserts to be ignored.
var NoDebug bool

func die(what string) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		panic(what)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", what)
		fmt.Fprintf(os.Stderr, "\t%s line %d\n", file, line)
		os.Exit(1)
	}
}

// Bool asserts that cond is false.
func Bool(cond bool) {
	if NoDebug {
		return
	}

	if !cond {
		die("assert.Bool failed")
	}
}

// Error asserts that err is nil.
func Error(err error) {
	if NoDebug {
		return
	}

	if nil != err {
		die(err.Error())
	}
}

// Error2 asserts that the actual error is the expected error.
func Error2(expected, actual error) {
	if NoDebug || (expected == actual) {
		return
	}

	if expected == nil {
		die(fmt.Sprintf("assert.Error2: %s", actual.Error()))
	}

	var should string
	if actual == nil {
		should = "no error was returned"
	} else {
		should = fmt.Sprintf("have '%s'", actual)
	}

	die(fmt.Sprintf("assert.Error2: expected '%s', but %s", expected, should))
}

// BoolT checks a boolean condition, calling Fatal on t if it is
// false.
func BoolT(t *testing.T, cond bool) {
	if !cond {
		t.Fatal("assert.Bool failed")
	}
}

// ErrorT checks whether the error is nil, calling Fatal on t if it
// isn't.
func ErrorT(t *testing.T, err error) {
	if nil != err {
		t.Fatalf("%s", err)
	}
}

// Error2T compares a pair of errors, calling Fatal on it if they
// don't match.
func Error2T(t *testing.T, expected, actual error) {
	if NoDebug || (expected == actual) {
		return
	}

	if expected == nil {
		die(fmt.Sprintf("assert.Error2: %s", actual.Error()))
	}

	var should string
	if actual == nil {
		should = "no error was returned"
	} else {
		should = fmt.Sprintf("have '%s'", actual)
	}

	die(fmt.Sprintf("assert.Error2: expected '%s', but %s", expected, should))
}
