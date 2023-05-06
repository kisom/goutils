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

// InvalidPEMType is used to indicate that we were expecting one type of PEM
// file, but saw another.
type InvalidPEMType struct {
	have string
	want []string
}

func (err *InvalidPEMType) Error() string {
	if len(err.want) == 1 {
		return fmt.Sprintf("invalid PEM type: have %s, expected %s", err.have, err.want[0])
	} else {
		return fmt.Sprintf("invalid PEM type: have %s, expected one of %s", err.have, strings.Join(err.want, ", "))
	}
}

// ErrInvalidPEMType returns a new InvalidPEMType error.
func ErrInvalidPEMType(have string, want ...string) error {
	return &InvalidPEMType{
		have: have,
		want: want,
	}
}

func LoadingError(t ErrorSourceType, err error) error {
	return fmt.Errorf("failed to load %s from disk: %w", t, err)
}

func ParsingError(t ErrorSourceType, err error) error {
	return fmt.Errorf("failed to parse %s: %w", t, err)
}

func DecodeError(t ErrorSourceType, err error) error {
	return fmt.Errorf("failed to decode %s: %w", t, err)
}

func VerifyError(t ErrorSourceType, err error) error {
	return fmt.Errorf("failed to verify %s: %w", t, err)
}

var ErrEncryptedPrivateKey = errors.New("private key is encrypted")
