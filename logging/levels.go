package logging

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// A Level represents a logging level.
type Level uint8

// The following constants represent logging levels in increasing levels of seriousness.
const (
	// LevelDebug are debug output useful during program testing
	// and debugging.
	LevelDebug = iota

	// LevelInfo is used for informational messages.
	LevelInfo

	// LevelNotice is for messages that are normal but
	// significant.
	LevelNotice

	// LevelWarning is for messages that are warning conditions:
	// they're not indicative of a failure, but of a situation
	// that may lead to a failure later.
	LevelWarning

	// LevelError is for messages indicating an error of some
	// kind.
	LevelError

	// LevelCritical are messages for critical conditions.
	LevelCritical

	// LevelAlert are for messages indicating that action
	// must be taken immediately.
	LevelAlert

	// LevelFatal messages are akin to syslog's LOG_EMERG: the
	// system is unusable and cannot continue execution.
	LevelFatal
)

// Cheap integer to fixed-width decimal ASCII.  Give a negative width
// to avoid zero-padding. (From log/log.go in the standard library).
func itoa(i int, wid int) string {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	return string(b[bp:])
}

func writeToOut(level Level) bool {
	if level < LevelWarning {
		return true
	}
	return false
}

var levelPrefix = [...]string{
	LevelDebug:    "[DEBUG] ",
	LevelInfo:     "[INFO] ",
	LevelNotice:   "[NOTICE] ",
	LevelWarning:  "[WARNING] ",
	LevelError:    "[ERROR] ",
	LevelCritical: "[CRITICAL] ",
	LevelAlert:    "[ALERT] ",
	LevelFatal:    "[FATAL] ",
}

var DateFormat = "2006-01-02T15:03:04-0700"

func (l *Logger) outputf(level Level, format string, v []interface{}) {
	if !l.Enabled() {
		return
	}

	if level >= l.level {
		domain := l.domain
		if level == LevelDebug {
			_, file, line, ok := runtime.Caller(2)
			if ok {
				domain += " " + file + ":" + itoa(line, -1)
			}
		}

		format = fmt.Sprintf("%s %s: %s%s\n",
			time.Now().Format(DateFormat),
			domain, levelPrefix[level], format)
		if writeToOut(level) {
			fmt.Fprintf(l.out, format, v...)
		} else {
			fmt.Fprintf(l.err, format, v...)
		}
	}
}

func (l *Logger) output(level Level, v []interface{}) {
	if !l.Enabled() {
		return
	}

	if level >= l.level {
		domain := l.domain
		if level == LevelDebug {
			_, file, line, ok := runtime.Caller(2)
			if ok {
				domain += " " + file + ":" + itoa(line, -1)
			}
		}

		format := fmt.Sprintf("%s %s: %s",
			time.Now().Format(DateFormat),
			domain, levelPrefix[level])
		if writeToOut(level) {
			fmt.Fprintf(l.out, format)
			fmt.Fprintln(l.out, v...)
		} else {
			fmt.Fprintf(l.err, format)
			fmt.Fprintln(l.err, v...)
		}
	}
}

// Fatalf logs a formatted message at the "fatal" level and then exits. The
// arguments are handled in the same manner as fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.outputf(LevelFatal, format, v)
	os.Exit(1)
}

// Fatal logs its arguments at the "fatal" level and then exits.
func (l *Logger) Fatal(v ...interface{}) {
	l.output(LevelFatal, v)
	os.Exit(1)
}

// Alertf logs a formatted message at the "alert" level. The
// arguments are handled in the same manner as fmt.Printf.
func (l *Logger) Alertf(format string, v ...interface{}) {
	l.outputf(LevelAlert, format, v)
}

// Alert logs its arguments at the "alert" level.
func (l *Logger) Alert(v ...interface{}) {
	l.output(LevelAlert, v)
}

// Criticalf logs a formatted message at the "critical" level. The
// arguments are handled in the same manner as fmt.Printf.
func (l *Logger) Criticalf(format string, v ...interface{}) {
	l.outputf(LevelCritical, format, v)
}

// Critical logs its arguments at the "critical" level.
func (l *Logger) Critical(v ...interface{}) {
	l.output(LevelCritical, v)
}

// Errorf logs a formatted message at the "error" level. The arguments
// are handled in the same manner as fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.outputf(LevelError, format, v)
}

// Error logs its arguments at the "error" level.
func (l *Logger) Error(v ...interface{}) {
	l.output(LevelError, v)
}

// Warningf logs a formatted message at the "warning" level. The
// arguments are handled in the same manner as fmt.Printf.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.outputf(LevelWarning, format, v)
}

// Warning logs its arguments at the "warning" level.
func (l *Logger) Warning(v ...interface{}) {
	l.output(LevelWarning, v)
}

// Noticef logs a formatted message at the "notice" level. The arguments
// are handled in the same manner as fmt.Printf.
func (l *Logger) Noticef(format string, v ...interface{}) {
	l.outputf(LevelNotice, format, v)
}

// Notice logs its arguments at the "notice" level.
func (l *Logger) Notice(v ...interface{}) {
	l.output(LevelNotice, v)
}

// Infof logs a formatted message at the "info" level. The arguments
// are handled in the same manner as fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.outputf(LevelInfo, format, v)
}

// Info logs its arguments at the "info" level.
func (l *Logger) Info(v ...interface{}) {
	l.output(LevelInfo, v)
}

// Debugf logs a formatted message at the "debug" level. The arguments
// are handled in the same manner as fmt.Printf. Note that debug
// logging will print the current
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.outputf(LevelDebug, format, v)
}

// Debug logs its arguments at the "debug" level.
func (l *Logger) Debug(v ...interface{}) {
	l.output(LevelDebug, v)
}

// Printf prints a formatted message at the default level.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.outputf(DefaultLevel, format, v)
}

// Print prints its arguments at the default level.
func (l *Logger) Print(v ...interface{}) {
	l.output(DefaultLevel, v)
}
