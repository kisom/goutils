package tee_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	tee "git.wntrmute.dev/kyle/goutils/tee"
)

// captureStdout redirects os.Stdout for the duration of fn and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	// Close writer to unblock reader and restore stdout
	_ = w.Close()
	b, _ := io.ReadAll(r)
	_ = r.Close()
	return string(b)
}

func TestNewOutEmpty_WritesToStdoutOnly(t *testing.T) {
	teeInst, err := tee.NewOut("")
	if err != nil {
		t.Fatalf("NewOut: %v", err)
	}

	out := captureStdout(t, func() {
		var n int
		if n, err = teeInst.Write([]byte("abc")); err != nil || n != 3 {
			t.Fatalf("Write got n=%d err=%v", n, err)
		}

		if n, err = teeInst.Printf("-%d-", 7); err != nil || n != len("-7-") {
			t.Fatalf("Printf got n=%d err=%v", n, err)
		}
	})

	if out != "abc-7-" {
		t.Fatalf("stdout = %q, want %q", out, "abc-7-")
	}
}

func TestNewOutWithFile_WritesToBoth(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "log.txt")

	teeInst, err := tee.NewOut(logPath)
	if err != nil {
		t.Fatalf("NewOut: %v", err)
	}
	defer func() { _ = teeInst.Close() }()

	out := captureStdout(t, func() {
		if _, err = teeInst.Write([]byte("x")); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if _, err = teeInst.Printf("%s", "y"); err != nil {
			t.Fatalf("Printf: %v", err)
		}
	})

	if out != "xy" {
		t.Fatalf("stdout = %q, want %q", out, "xy")
	}

	// Close to flush and release the file before reading
	if err = teeInst.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "xy" {
		t.Fatalf("file content = %q, want %q", string(data), "xy")
	}
}

func TestVPrintf_VerboseToggle(t *testing.T) {
	teeInst := &tee.Tee{} // stdout only

	out := captureStdout(t, func() {
		if n, err := teeInst.VPrintf("hello"); err != nil || n != 0 {
			t.Fatalf("VPrintf (quiet) got n=%d err=%v", n, err)
		}
	})
	if out != "" {
		t.Fatalf("stdout = %q, want empty when not verbose", out)
	}

	teeInst.Verbose = true
	out = captureStdout(t, func() {
		if n, err := teeInst.VPrintf("%s", "hello"); err != nil || n != len("hello") {
			t.Fatalf("VPrintf (verbose) got n=%d err=%v", n, err)
		}
	})
	if out != "hello" {
		t.Fatalf("stdout = %q, want %q", out, "hello")
	}
}

func TestWrite_StdoutErrorDoesNotWriteToFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "log.txt")
	teeInst, err := tee.NewOut(logPath)
	if err != nil {
		t.Fatalf("NewOut: %v", err)
	}
	defer func() { _ = teeInst.Close() }()

	// Replace stdout with a closed pipe writer to force write error.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	_ = w.Close() // immediately close to cause EPIPE on write
	defer func() {
		os.Stdout = old
		_ = r.Close()
	}()

	var n int
	if n, err = teeInst.Write([]byte("abc")); err == nil {
		t.Fatalf("expected error writing to closed stdout, got n=%d err=nil", n)
	}

	// Ensure file remained empty because stdout write failed first.
	_ = teeInst.Close()
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("file content = %q, want empty due to stdout failure", string(data))
	}
}

func TestGlobal_OpenPrintfVPrintfClose(t *testing.T) {
	// Ensure a clean slate for global tee
	_ = tee.Close()
	tee.SetVerbose(false)

	dir := t.TempDir()
	logPath := filepath.Join(dir, "glog.txt")

	if err := tee.Open(logPath); err != nil {
		t.Fatalf("Open: %v", err)
	}

	out := captureStdout(t, func() {
		if _, err := tee.Printf("A"); err != nil {
			t.Fatalf("Printf: %v", err)
		}
		// Not verbose yet, should not print
		if n, err := tee.VPrintf("B"); err != nil || n != 0 {
			t.Fatalf("VPrintf (quiet) n=%d err=%v", n, err)
		}
		tee.SetVerbose(true)
		if _, err := tee.VPrintf("C%d", 1); err != nil {
			t.Fatalf("VPrintf (verbose): %v", err)
		}
	})

	if out != "AC1" {
		t.Fatalf("stdout = %q, want %q", out, "AC1")
	}

	if err := tee.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "AC1" {
		t.Fatalf("file content = %q, want %q", string(data), "AC1")
	}

	// Reset global tee for other tests/packages
	_ = tee.Close()
	tee.SetVerbose(false)
}
