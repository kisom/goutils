package msg_test

import (
	"bytes"
	"testing"

	"git.wntrmute.dev/kyle/goutils/msg"
)

func checkExpected(buf *bytes.Buffer, expected string) bool {
	return buf.String() == expected
}

func resetBuf() *bytes.Buffer {
	buf := &bytes.Buffer{}
	msg.SetWriter(buf)

	return buf
}

func TestVerbosePrint(t *testing.T) {
	buf := resetBuf()

	msg.SetVerbose(false) // ensure verbose is explicitly not set

	msg.Vprint("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Vprintf("hello, %s", "world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Vprintln("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.SetVerbose(true)
	msg.Vprint("hello, world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Vprintf("hello, %s", "world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Vprintln("hello, world")
	if !checkExpected(buf, "hello, world\n") {
		t.Fatalf("expected output %q, have %q", "hello, world\n", buf.String())
	}
}

func TestQuietPrint(t *testing.T) {
	buf := resetBuf()

	msg.SetQuiet(true)

	msg.Qprint("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Qprintf("hello, %s", "world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Qprintln("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.SetQuiet(false)
	msg.Qprint("hello, world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Qprintf("hello, %s", "world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Qprintln("hello, world")
	if !checkExpected(buf, "hello, world\n") {
		t.Fatalf("expected output %q, have %q", "hello, world\n", buf.String())
	}
}

func TestDebugPrint(t *testing.T) {
	buf := resetBuf()

	msg.SetDebug(false) // ensure debug is explicitly not set

	msg.Dprint("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Dprintf("hello, %s", "world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.Dprintln("hello, world")
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.StackTrace()
	if buf.Len() != 0 {
		t.Fatalf("expected no output, have %s", buf.String())
	}

	msg.SetDebug(true)
	msg.Dprint("hello, world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Dprintf("hello, %s", "world")
	if !checkExpected(buf, "hello, world") {
		t.Fatalf("expected output %q, have %q", "hello, world", buf.String())
	}
	buf.Reset()

	msg.Dprintln("hello, world")
	if !checkExpected(buf, "hello, world\n") {
		t.Fatalf("expected output %q, have %q", "hello, world\n", buf.String())
	}
	buf.Reset()

	msg.StackTrace()
	if buf.Len() == 0 {
		t.Fatal("expected stack trace output, received no output")
	}
}
