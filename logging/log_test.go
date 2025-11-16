package logging_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"git.wntrmute.dev/kyle/goutils/logging"
)

// A list of implementations that should be tested.
var implementations []logging.Logger

func init() {
	lw := logging.NewLogWriter(&bytes.Buffer{}, nil)
	cw := logging.NewConsole()

	implementations = append(implementations, lw)
	implementations = append(implementations, cw)
}

func TestFileSetup(t *testing.T) {
	fw1, err := logging.NewFile("fw1.log", true)
	if err != nil {
		t.Fatalf("failed to create new file logger: %v", err)
	}

	fw2, err := logging.NewSplitFile("fw2.log", "fw2.err", true)
	if err != nil {
		t.Fatalf("failed to create new split file logger: %v", err)
	}

	implementations = append(implementations, fw1)
	implementations = append(implementations, fw2)
}

func TestImplementations(_ *testing.T) {
	for _, l := range implementations {
		l.Info("TestImplementations", "Info message",
			map[string]string{"type": fmt.Sprintf("%T", l)})
		l.Warn("TestImplementations", "Warning message",
			map[string]string{"type": fmt.Sprintf("%T", l)})
	}
}

func TestCloseLoggers(t *testing.T) {
	for _, l := range implementations {
		if err := l.Close(); err != nil {
			t.Errorf("failed to close logger: %v", err)
		}
	}
}

func TestDestroyLogFiles(t *testing.T) {
	if err := os.Remove("fw1.log"); err != nil {
		t.Errorf("failed to remove fw1.log: %v", err)
	}

	if err := os.Remove("fw2.log"); err != nil {
		t.Errorf("failed to remove fw2.log: %v", err)
	}

	if err := os.Remove("fw2.err"); err != nil {
		t.Errorf("failed to remove fw2.err: %v", err)
	}
}

func TestMulti(t *testing.T) {
	c1 := logging.NewConsole()
	c2 := logging.NewConsole()
	m := logging.NewMulti(c1, c2)
	if !m.Good() {
		t.Fatal("failed to set up multi logger")
	}
}
