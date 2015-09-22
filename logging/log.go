// Package logging provides domain-based logging in the same style as
// sylog. Domains are some name for which logging can be selectively
// enabled or disabled. Logging also differentiates between normal
// messages (which are sent to standard output) and errors, which are
// sent to standard error; debug messages will also include the file
// and line number.
//
// Domains are intended for identifying logging subystems. A domain
// can be suppressed with Suppress, and re-enabled with Enable. There
// are prefixed versions of these as well.
//
// This package was adapted from the CFSSL logging code.
package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kisom/goutils/mwc"
)

var logConfig = struct {
	registered map[string]*Logger
	lock       *sync.Mutex
}{
	registered: map[string]*Logger{},
	lock:       new(sync.Mutex),
}

// DefaultLevel defaults to the notice level of logging.
const DefaultLevel = LevelNotice

// Init returns a new default logger. The domain is set to the
// program's name, and the default logging level is used.
func Init() *Logger {
	l, _ := New(filepath.Base(os.Args[0]), DefaultLevel)
	return l
}

// A Logger writes logs on behalf of a particular domain at a certain
// level.
type Logger struct {
	enabled bool
	lock    *sync.Mutex
	domain  string
	level   Level
	out     io.WriteCloser
	err     io.WriteCloser
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
	l, ok := logConfig.registered[domain]
	if ok {
		l.Suppress()
	}
}

// SuppressPrefix suppress logs whose domain is prefixed with the
// prefix.
func SuppressPrefix(prefix string) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for domain, l := range logConfig.registered {
		if strings.HasPrefix(domain, prefix) {
			l.Suppress()
		}
	}
}

// SuppressAll suppresses all logging output.
func SuppressAll() {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for _, l := range logConfig.registered {
		l.Suppress()
	}
}

// Enable enables logs from a specific domain.
func Enable(domain string) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	l, ok := logConfig.registered[domain]
	if ok {
		l.Enable()
	}
}

// EnablePrefix enables logs whose domain is prefixed with prefix.
func EnablePrefix(prefix string) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for domain, l := range logConfig.registered {
		if strings.HasPrefix(domain, prefix) {
			l.Enable()
		}
	}
}

// EnableAll enables all domains.
func EnableAll() {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()
	for _, l := range logConfig.registered {
		l.Enable()
	}
}

// New returns a new logger that writes to standard output for Notice
// and below and standard error for levels above Notice. If a logger
// with the same domain exists, the logger will set its level to level
// and return the logger; in this case, the registered return value
// will be true.
func New(domain string, level Level) (l *Logger, registered bool) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()

	l = logConfig.registered[domain]
	if l != nil {
		l.SetLevel(level)
		return l, true
	}

	l = &Logger{
		domain: domain,
		level:  level,
		out:    os.Stdout,
		err:    os.Stderr,
		lock:   new(sync.Mutex),
	}

	l.Enable()
	logConfig.registered[domain] = l
	return l, false
}

// NewWriters returns a new logger that writes to the w io.WriteCloser
// for Notice and below and to the e io.WriteCloser for levels above
// Notice. If e is nil, w will be used. If a logger with the same
// domain exists, the logger will set its level to level and return
// the logger; in this case, the registered return value will be true.
func NewFromWriters(domain string, level Level, w, e io.WriteCloser) (l *Logger, registered bool) {
	logConfig.lock.Lock()
	defer logConfig.lock.Unlock()

	l = logConfig.registered[domain]
	if l != nil {
		l.SetLevel(level)
		return l, true
	}

	if w == nil {
		w = os.Stdout
	}

	if e == nil {
		e = w
	}

	l = &Logger{
		domain: domain,
		level:  level,
		out:    w,
		err:    e,
		lock:   new(sync.Mutex),
	}

	l.Enable()
	logConfig.registered[domain] = l
	return l, false
}

// NewFile returns a new logger that opens the files for writing. If
// multiplex is true, output will be multiplexed to standard output
// and standard error as well.
func NewFromFile(domain string, level Level, outFile, errFile string, multiplex bool, flags int) (*Logger, error) {
	l := &Logger{
		domain: domain,
		level:  level,
		lock:   new(sync.Mutex),
	}

	outf, err := os.OpenFile(outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY|flags, 0644)
	if err != nil {
		return nil, err
	}

	errf, err := os.OpenFile(errFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY|flags, 0644)
	if err != nil {
		return nil, err
	}

	if multiplex {
		l.out = mwc.MultiWriteCloser(outf, os.Stdout)
		l.err = mwc.MultiWriteCloser(errf, os.Stderr)
	} else {
		l.out = outf
		l.err = errf
	}

	Enable(domain)
	return l, nil
}

// Enable allows output from the logger.
func (l *Logger) Enable() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.enabled = true
}

// Enabled returns true if the logger is enabled.
func (l *Logger) Enabled() bool {
	return l.enabled
}

// Suppress ignores output from the logger.
func (l *Logger) Suppress() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.enabled = false
}

// Domain returns the domain of the logger.
func (l *Logger) Domain() string {
	return l.domain
}

// SetLevel changes the level of the logger.
func (l *Logger) SetLevel(level Level) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}
