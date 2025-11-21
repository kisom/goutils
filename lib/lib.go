package lib

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// IsDigit checks if a byte is a decimal digit.
func IsDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

const signedaMask64 = 1<<63 - 1

// ParseDuration parses a duration string into a time.Duration.
// It supports standard units (ns, us/µs, ms, s, m, h) plus extended units:
// d (days, 24h), w (weeks, 7d), y (years, 365d).
// Units can be combined without spaces, e.g., "1y2w3d4h5m6s".
// Case-insensitive. Years and days are approximations (no leap seconds/months).
// Returns an error for invalid input.
func ParseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(s) // Normalize to lowercase for case-insensitivity.
	if s == "" {
		return 0, errors.New("empty duration string")
	}

	var total time.Duration
	i := 0
	for i < len(s) {
		// Parse the number part.
		start := i
		for i < len(s) && IsDigit(s[i]) {
			i++
		}
		if start == i {
			return 0, fmt.Errorf("expected number at position %d", start)
		}
		numStr := s[start:i]
		num, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number %q: %w", numStr, err)
		}

		// Parse the unit part.
		if i >= len(s) {
			return 0, fmt.Errorf("expected unit after number %q", numStr)
		}
		unitStart := i
		i++ // Consume the first char of the unit.
		unit := s[unitStart:i]

		// Handle potential two-char units like "ms".
		if unit == "m" && i < len(s) && s[i] == 's' {
			i++ // Consume the 's'.
			unit = "ms"
		}

		// Convert to duration based on unit.
		var d time.Duration
		switch unit {
		case "ns":
			d = time.Nanosecond * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "us", "µs":
			d = time.Microsecond * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "ms":
			d = time.Millisecond * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "s":
			d = time.Second * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "m":
			d = time.Minute * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "h":
			d = time.Hour * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "d":
			d = 24 * time.Hour * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "w":
			d = 7 * 24 * time.Hour * time.Duration(num&signedaMask64) // #nosec G115 - masked off
		case "y":
			// Approximate, non-leap year.
			d = 365 * 24 * time.Hour * time.Duration(num&signedaMask64) // #nosec G115 - masked off;
		default:
			return 0, fmt.Errorf("unknown unit %q at position %d", s[unitStart:i], unitStart)
		}

		total += d
	}

	return total, nil
}

type HexEncodeMode uint8

const (
	// HexEncodeLower prints the bytes as lowercase hexadecimal.
	HexEncodeLower HexEncodeMode = iota + 1
	// HexEncodeUpper prints the bytes as uppercase hexadecimal.
	HexEncodeUpper
	// HexEncodeLowerColon prints the bytes as lowercase hexadecimal
	// with colons between each pair of bytes.
	HexEncodeLowerColon
	// HexEncodeUpperColon prints the bytes as uppercase hexadecimal
	// with colons between each pair of bytes.
	HexEncodeUpperColon
	// HexEncodeBytes prints the string as a sequence of []byte.
	HexEncodeBytes
	// HexEncodeBase64 prints the string as a base64-encoded string.
	HexEncodeBase64
)

func (m HexEncodeMode) String() string {
	switch m {
	case HexEncodeLower:
		return "lower"
	case HexEncodeUpper:
		return "upper"
	case HexEncodeLowerColon:
		return "lcolon"
	case HexEncodeUpperColon:
		return "ucolon"
	case HexEncodeBytes:
		return "bytes"
	case HexEncodeBase64:
		return "base64"
	default:
		panic("invalid hex encode mode")
	}
}

func ParseHexEncodeMode(s string) HexEncodeMode {
	switch strings.ToLower(s) {
	case "lower":
		return HexEncodeLower
	case "upper":
		return HexEncodeUpper
	case "lcolon":
		return HexEncodeLowerColon
	case "ucolon":
		return HexEncodeUpperColon
	case "bytes":
		return HexEncodeBytes
	case "base64":
		return HexEncodeBase64
	}

	panic("invalid hex encode mode")
}

func hexColons(s string) string {
	if len(s)%2 != 0 {
		fmt.Fprintf(os.Stderr, "hex string: %s\n", s)
		fmt.Fprintf(os.Stderr, "hex length: %d\n", len(s))
		panic("invalid hex string length")
	}

	n := len(s)
	if n <= 2 {
		return s
	}

	pairCount := n / 2
	if n%2 != 0 {
		pairCount++
	}

	var b strings.Builder
	b.Grow(n + pairCount - 1)

	for i := 0; i < n; i += 2 {
		b.WriteByte(s[i])

		if i+1 < n {
			b.WriteByte(s[i+1])
		}

		if i+2 < n {
			b.WriteByte(':')
		}
	}

	return b.String()
}

func hexEncode(b []byte) string {
	s := hex.EncodeToString(b)

	if len(s)%2 != 0 {
		s = "0" + s
	}

	return s
}

func bytesAsByteSliceString(buf []byte) string {
	sb := &strings.Builder{}
	sb.WriteString("[]byte{")
	for i := range buf {
		fmt.Fprintf(sb, "0x%02x, ", buf[i])
	}
	sb.WriteString("}")

	return sb.String()
}

// HexEncode encodes the given bytes as a hexadecimal string. It
// also supports a few other binary-encoding formats as well.
func HexEncode(b []byte, mode HexEncodeMode) string {
	switch mode {
	case HexEncodeLower:
		return hexEncode(b)
	case HexEncodeUpper:
		return strings.ToUpper(hexEncode(b))
	case HexEncodeLowerColon:
		return hexColons(hexEncode(b))
	case HexEncodeUpperColon:
		return strings.ToUpper(hexColons(hexEncode(b)))
	case HexEncodeBytes:
		return bytesAsByteSliceString(b)
	case HexEncodeBase64:
		return base64.StdEncoding.EncodeToString(b)
	default:
		panic("invalid hex encode mode")
	}
}

// DummyWriteCloser wraps an io.Writer in a struct with a no-op Close.
type DummyWriteCloser struct {
	w io.Writer
}

func (dwc *DummyWriteCloser) Write(p []byte) (int, error) {
	return dwc.w.Write(p)
}

func (dwc *DummyWriteCloser) Close() error {
	return nil
}
