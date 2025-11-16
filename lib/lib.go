// Package lib contains functions useful for most programs.
package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var progname = filepath.Base(os.Args[0])

const (
	daysInYear        = 365
	digitWidth        = 10
	hoursInQuarterDay = 6
)

// ProgName returns what lib thinks the program name is, namely the
// basename of argv0.
//
// It is similar to the Linux __progname function.
func ProgName() string {
	return progname
}

// Warnx displays a formatted error message to standard error, à la
// warnx(3).
func Warnx(format string, a ...any) (int, error) {
	format = fmt.Sprintf("[%s] %s", progname, format)
	format += "\n"
	return fmt.Fprintf(os.Stderr, format, a...)
}

// Warn displays a formatted error message to standard output,
// appending the error string, à la warn(3).
func Warn(err error, format string, a ...any) (int, error) {
	format = fmt.Sprintf("[%s] %s", progname, format)
	format += ": %v\n"
	a = append(a, err)
	return fmt.Fprintf(os.Stderr, format, a...)
}

// Errx displays a formatted error message to standard error and exits
// with the status code from `exit`, à la errx(3).
func Errx(exit int, format string, a ...any) {
	format = fmt.Sprintf("[%s] %s", progname, format)
	format += "\n"
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(exit)
}

// Err displays a formatting error message to standard error,
// appending the error string, and exits with the status code from
// `exit`, à la err(3).
func Err(exit int, err error, format string, a ...any) {
	format = fmt.Sprintf("[%s] %s", progname, format)
	format += ": %v\n"
	a = append(a, err)
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(exit)
}

// Itoa provides cheap integer to fixed-width decimal ASCII.  Give a
// negative width to avoid zero-padding. Adapted from the 'itoa'
// function in the log/log.go file in the standard library.
func Itoa(i int, wid int) string {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= digitWidth || wid > 1 {
		wid--
		q := i / digitWidth
		b[bp] = byte('0' + i - q*digitWidth)
		bp--
		i = q
	}

	b[bp] = byte('0' + i)
	return string(b[bp:])
}

var (
	dayDuration  = 24 * time.Hour
	yearDuration = (daysInYear * dayDuration) + (hoursInQuarterDay * time.Hour)
)

// Duration returns a prettier string for time.Durations.
func Duration(d time.Duration) string {
	var s string
	if d >= yearDuration {
		years := int64(d / yearDuration)
		s += fmt.Sprintf("%dy", years)
		d -= time.Duration(years) * yearDuration
	}

	if d >= dayDuration {
		days := d / dayDuration
		s += fmt.Sprintf("%dd", days)
	}

	if s != "" {
		return s
	}

	d %= 1 * time.Second
	hours := int64(d / time.Hour)
	d -= time.Duration(hours) * time.Hour
	s += fmt.Sprintf("%dh%s", hours, d)
	return s
}
