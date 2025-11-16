package certerr

import (
	"errors"
	"strings"
	"testing"
)

func TestTypedErrorWrappingAndFormatting(t *testing.T) {
	cause := errors.New("bad data")
	err := DecodeError(ErrorSourceCertificate, cause)

	// Ensure we can retrieve the typed error
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("expected errors.As to retrieve *certerr.Error, got %T", err)
	}
	if e.Kind != KindDecode {
		t.Fatalf("unexpected kind: %v", e.Kind)
	}
	if e.Source != ErrorSourceCertificate {
		t.Fatalf("unexpected source: %v", e.Source)
	}

	// Check message format (no trailing punctuation enforced by content)
	msg := e.Error()
	if !strings.Contains(msg, "failed to decode certificate") || !strings.Contains(msg, "bad data") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestErrorsIsOnWrappedSentinel(t *testing.T) {
	err := DecodeError(ErrorSourcePrivateKey, ErrEncryptedPrivateKey)
	if !errors.Is(err, ErrEncryptedPrivateKey) {
		t.Fatalf("expected errors.Is to match ErrEncryptedPrivateKey")
	}
}

func TestInvalidPEMTypeMessageSingle(t *testing.T) {
	err := ErrInvalidPEMType("FOO", "CERTIFICATE")
	want := "invalid PEM type: have FOO, expected CERTIFICATE"
	if err.Error() != want {
		t.Fatalf("unexpected error message: got %q, want %q", err.Error(), want)
	}
}

func TestInvalidPEMTypeMessageMultiple(t *testing.T) {
	err := ErrInvalidPEMType("FOO", "CERTIFICATE", "NEW CERTIFICATE REQUEST")
	if !strings.Contains(
		err.Error(),
		"invalid PEM type: have FOO, expected one of CERTIFICATE, NEW CERTIFICATE REQUEST",
	) {
		t.Fatalf("unexpected error message: %q", err.Error())
	}
}
