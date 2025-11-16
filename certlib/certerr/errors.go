package certerr

import (
	"errors"
	"fmt"
	"strings"
)

// ErrEmptyCertificate indicates that a certificate could not be processed
// because there was no data to process.
var ErrEmptyCertificate = errors.New("certlib: empty certificate")

type ErrorSourceType uint8

func (t ErrorSourceType) String() string {
	switch t {
	case ErrorSourceCertificate:
		return "certificate"
	case ErrorSourcePrivateKey:
		return "private key"
	case ErrorSourceCSR:
		return "CSR"
	case ErrorSourceSCTList:
		return "SCT list"
	case ErrorSourceKeypair:
		return "TLS keypair"
	default:
		panic(fmt.Sprintf("unknown error source %d", t))
	}
}

const (
	ErrorSourceCertificate ErrorSourceType = 1
	ErrorSourcePrivateKey  ErrorSourceType = 2
	ErrorSourceCSR         ErrorSourceType = 3
	ErrorSourceSCTList     ErrorSourceType = 4
	ErrorSourceKeypair     ErrorSourceType = 5
)

// ErrorKind is a broad classification describing what went wrong.
type ErrorKind uint8

const (
	KindParse ErrorKind = iota + 1
	KindDecode
	KindVerify
	KindLoad
)

func (k ErrorKind) String() string {
	switch k {
	case KindParse:
		return "parse"
	case KindDecode:
		return "decode"
	case KindVerify:
		return "verify"
	case KindLoad:
		return "load"
	default:
		return "unknown"
	}
}

// Error is a typed, wrapped error with structured context for programmatic checks.
// It implements error and supports errors.Is/As via Unwrap.
type Error struct {
	Source ErrorSourceType // which domain produced the error (certificate, private key, etc.)
	Kind   ErrorKind       // operation category (parse, decode, verify, load)
	Op     string          // optional operation or function name
	Err    error           // wrapped cause
}

func (e *Error) Error() string {
	// Keep message format consistent with existing helpers: "failed to <kind> <source>: <err>"
	// Do not include Op by default to preserve existing output expectations.
	return fmt.Sprintf("failed to %s %s: %v", e.Kind.String(), e.Source.String(), e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// InvalidPEMTypeError is used to indicate that we were expecting one type of PEM
// file, but saw another.
type InvalidPEMTypeError struct {
	have string
	want []string
}

func (err *InvalidPEMTypeError) Error() string {
	if len(err.want) == 1 {
		return fmt.Sprintf("invalid PEM type: have %s, expected %s", err.have, err.want[0])
	}
	return fmt.Sprintf("invalid PEM type: have %s, expected one of %s", err.have, strings.Join(err.want, ", "))
}

// ErrInvalidPEMType returns a new InvalidPEMTypeError error.
func ErrInvalidPEMType(have string, want ...string) error {
	return &InvalidPEMTypeError{
		have: have,
		want: want,
	}
}

func LoadingError(t ErrorSourceType, err error) error {
	return &Error{Source: t, Kind: KindLoad, Err: err}
}

func ParsingError(t ErrorSourceType, err error) error {
	return &Error{Source: t, Kind: KindParse, Err: err}
}

func DecodeError(t ErrorSourceType, err error) error {
	return &Error{Source: t, Kind: KindDecode, Err: err}
}

func VerifyError(t ErrorSourceType, err error) error {
	return &Error{Source: t, Kind: KindVerify, Err: err}
}

var ErrEncryptedPrivateKey = errors.New("private key is encrypted")
