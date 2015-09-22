// Package logging is an adaptation of the CFSSL logging library. It
// operates on domains, which are components for which logging can be
// selectively enabled or disabled. It also differentiates between
// normal messages (which are sent to standard output) and errors,
// which are sent to standard error.
package logging

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/kisom/goutils/mwc"
)

var logConfig = struct {
	registered map[string]bool
	lock       *sync.Mutex
}{
	registered: map[string]bool{},
	lock:       new(sync.Mutex),
}

// DefaultLevel defaults to the notice level of logging.
const DefaultLevel = LevelNotice

// Init returns a new default logger. The domain is set to the
// program's name, and the default logging level is used.
func Init() *Logger {
	return New(filepath.Base(os.Args[0]), DefaultLevel)
}

// A Logger writes logs on behalf of a particular domain at a certain
// level.
type Logger struct {
	domain string
	level  Level
	out    io.WriteCloser
	err    io.WriteCloser
}

// Close closes the log's writers and suppresses the logger.
func (l *Logger) Close() error {
	Suppress(l.domain)
	err := l.out.Close()
	if err != nil {
		return nil
	}

	return l.err.Close()
}

// Suppress ignores logs from a specific domain.
func Suppress(domain string) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	logConfig.registered[domain] = false
}

// SuppressAll suppresses all logging output.
func SuppressAll() {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for domain := range logConfig.registered {
		logConfig.registered[domain] = false
	}
}

// Enable enables logs from a specific domain.
func Enable(domain string) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	logConfig.registered[domain] = true
}

// EnableAll enables all domains.
func EnableAll() {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for domain := range logConfig.registered {
		logConfig.registered[domain] = true
	}
}

// New returns a new logger that writes to standard output for Notice
// and below and standard error for levels above Notice.
func New(domain string, level Level) *Logger {
	l := &Logger{
		domain: domain,
		level:  level,
		out:    os.Stdout,
		err:    os.Stderr,
	}

	Enable(domain)
	return l
}

// NewWriters returns a new logger that writes to the w io.WriteCloser for
// Notice and below and to the e io.WriteCloser for levels above Notice. If e is nil, w will be used.
func NewWriters(domain string, level Level, w, e io.WriteCloser) *Logger {
	if e == nil {
		e = w
	}

	l := &Logger{
		domain: domain,
		level:  level,
		out:    w,
		err:    e,
	}

	Enable(domain)
	return l
}

// NewFile returns a new logger that opens the files for writing. If
// multiplex is true, output will be multiplexed to standard output
// and standard error as well.
func NewFile(domain string, level Level, outFile, errFile string, multiplex bool) (*Logger, error) {
	l := &Logger{
		domain: domain,
		level:  level,
	}

	var err error
	l.out, err = os.OpenFile(outFile, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	l.err, err = os.OpenFile(errFile, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	if multiplex {
		l.out = mwc.MultiWriteCloser(l.out, os.Stdout)
		l.err = mwc.MultiWriteCloser(l.err, os.Stderr)
	}

	Enable(domain)
	return l, nil
}

// Enable allows output from the logger.
func (l *Logger) Enable() {
	Enable(l.domain)
}

// Suppress ignores output from the logger.
func (l *Logger) Suppress() {
	Suppress(l.domain)
}

// Domain returns the domain of the logger.
func (l *Logger) Domain() string {
	return l.domain
}

// SetLevel changes the level of the logger.
func (l *Logger) SetLevel(level Level) {
	l.level = level
}
