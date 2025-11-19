package lib_test

import (
	"testing"
	"time"

	"git.wntrmute.dev/kyle/goutils/lib"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Valid durations
		{"hour", "1h", time.Hour, false},
		{"day", "2d", 2 * 24 * time.Hour, false},
		{"minute", "3m", 3 * time.Minute, false},
		{"second", "4s", 4 * time.Second, false},

		// Edge cases
		{"zero seconds", "0s", 0, false},
		{"empty string", "", 0, true},
		{"no numeric before unit", "h", 0, true},
		{"invalid unit", "1x", 0, true},
		{"non-numeric input", "abc", 0, true},
		{"missing unit", "10", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lib.ParseDuration(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestHexEncode_LowerUpper(t *testing.T) {
	b := []byte{0x0f, 0xa1, 0x00, 0xff}

	gotLower := lib.HexEncode(b, lib.HexEncodeLower)
	if gotLower != "0fa100ff" {
		t.Fatalf("lib.HexEncode lower: expected %q, got %q", "0fa100ff", gotLower)
	}

	gotUpper := lib.HexEncode(b, lib.HexEncodeUpper)
	if gotUpper != "0FA100FF" {
		t.Fatalf("lib.HexEncode upper: expected %q, got %q", "0FA100FF", gotUpper)
	}
}

func TestHexEncode_ColonModes(t *testing.T) {
	// Includes leading zero nibble and a zero byte to verify padding and separators
	b := []byte{0x0f, 0xa1, 0x00, 0xff}

	gotLColon := lib.HexEncode(b, lib.HexEncodeLowerColon)
	if gotLColon != "0f:a1:00:ff" {
		t.Fatalf("lib.HexEncode colon lower: expected %q, got %q", "0f:a1:00:ff", gotLColon)
	}

	gotUColon := lib.HexEncode(b, lib.HexEncodeUpperColon)
	if gotUColon != "0F:A1:00:FF" {
		t.Fatalf("lib.HexEncode colon upper: expected %q, got %q", "0F:A1:00:FF", gotUColon)
	}
}

func TestHexEncode_EmptyInput(t *testing.T) {
	var b []byte
	if got := lib.HexEncode(b, lib.HexEncodeLower); got != "" {
		t.Fatalf("empty lower: expected empty string, got %q", got)
	}
	if got := lib.HexEncode(b, lib.HexEncodeUpper); got != "" {
		t.Fatalf("empty upper: expected empty string, got %q", got)
	}
	if got := lib.HexEncode(b, lib.HexEncodeLowerColon); got != "" {
		t.Fatalf("empty colon lower: expected empty string, got %q", got)
	}
	if got := lib.HexEncode(b, lib.HexEncodeUpperColon); got != "" {
		t.Fatalf("empty colon upper: expected empty string, got %q", got)
	}
}

func TestHexEncode_SingleByte(t *testing.T) {
	b := []byte{0x0f}
	if got := lib.HexEncode(b, lib.HexEncodeLower); got != "0f" {
		t.Fatalf("single byte lower: expected %q, got %q", "0f", got)
	}
	if got := lib.HexEncode(b, lib.HexEncodeUpper); got != "0F" {
		t.Fatalf("single byte upper: expected %q, got %q", "0F", got)
	}
	// For a single byte, colon modes should not introduce separators
	if got := lib.HexEncode(b, lib.HexEncodeLowerColon); got != "0f" {
		t.Fatalf("single byte colon lower: expected %q, got %q", "0f", got)
	}
	if got := lib.HexEncode(b, lib.HexEncodeUpperColon); got != "0F" {
		t.Fatalf("single byte colon upper: expected %q, got %q", "0F", got)
	}
}

func TestHexEncode_InvalidModePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid mode, but function returned normally")
		}
	}()
	// 0 is not a valid lib.HexEncodeMode (valid modes start at 1)
	_ = lib.HexEncode([]byte{0x01}, lib.HexEncodeMode(0))
}
