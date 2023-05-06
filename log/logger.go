// Package syslog is a syslog-type facility for logging.
package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	gsyslog "github.com/hashicorp/go-syslog"
)

type logger struct {
	l gsyslog.Syslogger
	p gsyslog.Priority
}

func (log *logger) printf(p gsyslog.Priority, format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	if p <= log.p {
		fmt.Printf("%s [%s] ", prioritiev[p], timestamp())
		fmt.Printf(format, args...)
	}

	if log.l != nil {
		log.l.WriteLevel(p, []byte(fmt.Sprintf(format, args...)))
	}
}

func (log *logger) print(p gsyslog.Priority, args ...interface{}) {
	if p <= log.p {
		fmt.Printf("%s [%s] ", prioritiev[p], timestamp())
		fmt.Print(args...)
	}

	if log.l != nil {
		log.l.WriteLevel(p, []byte(fmt.Sprint(args...)))
	}
}

func (log *logger) println(p gsyslog.Priority, args ...interface{}) {
	if p <= log.p {
		fmt.Printf("%s [%s] ", prioritiev[p], timestamp())
		fmt.Println(args...)
	}

	if log.l != nil {
		log.l.WriteLevel(p, []byte(fmt.Sprintln(args...)))
	}
}

func (log *logger) spew(args ...interface{}) {
	if log.p == gsyslog.LOG_DEBUG {
		spew.Dump(args...)
	}
}

func (log *logger) adjustPriority(level string) error {
	priority, ok := priorities[level]
	if !ok {
		return fmt.Errorf("log: unknown priority %s", level)
	}

	log.p = priority
	return nil
}

var log = &logger{p: gsyslog.LOG_WARNING}

var priorities = map[string]gsyslog.Priority{
	"EMERG":   gsyslog.LOG_EMERG,
	"ALERT":   gsyslog.LOG_ALERT,
	"CRIT":    gsyslog.LOG_CRIT,
	"ERR":     gsyslog.LOG_ERR,
	"WARNING": gsyslog.LOG_WARNING,
	"NOTICE":  gsyslog.LOG_NOTICE,
	"INFO":    gsyslog.LOG_INFO,
	"DEBUG":   gsyslog.LOG_DEBUG,
}

var prioritiev = map[gsyslog.Priority]string{
	gsyslog.LOG_EMERG:   "EMERG",
	gsyslog.LOG_ALERT:   "ALERT",
	gsyslog.LOG_CRIT:    "CRIT",
	gsyslog.LOG_ERR:     "ERR",
	gsyslog.LOG_WARNING: "WARNING",
	gsyslog.LOG_NOTICE:  "NOTICE",
	gsyslog.LOG_INFO:    "INFO",
	gsyslog.LOG_DEBUG:   "DEBUG",
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}

type Options struct {
	Level       string
	Tag         string
	Facility    string
	WriteSyslog bool
}

// DefaultOptions returns a sane set of defaults for syslog, using the program
// name as the tag name. withSyslog controls whether logs should be sent to
// syslog, too.
func DefaultOptions(tag string, withSyslog bool) *Options {
	if tag == "" {
		tag = os.Args[0]
	}

	return &Options{
		Level:       "WARNING",
		Tag:         tag,
		Facility:    "daemon",
		WriteSyslog: withSyslog,
	}
}

// DefaultDebugOptions returns a sane set of debug defaults for syslog,
// using the program name as the tag name. withSyslog controls whether logs
// should be sent to syslog, too.
func DefaultDebugOptions(tag string, withSyslog bool) *Options {
	if tag == "" {
		tag = os.Args[0]
	}

	return &Options{
		Level:       "DEBUG",
		Facility:    "daemon",
		WriteSyslog: withSyslog,
	}
}

func Setup(opts *Options) error {
	priority, ok := priorities[opts.Level]
	if !ok {
		return fmt.Errorf("log: unknown priority %s", opts.Level)
	}

	log.p = priority

	if opts.WriteSyslog {
		var err error
		log.l, err = gsyslog.NewLogger(priority, opts.Facility, opts.Tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func Debug(args ...interface{}) {
	log.print(gsyslog.LOG_DEBUG, args...)
}

func Info(args ...interface{}) {
	log.print(gsyslog.LOG_INFO, args...)
}

func Notice(args ...interface{}) {
	log.print(gsyslog.LOG_NOTICE, args...)
}

func Warning(args ...interface{}) {
	log.print(gsyslog.LOG_WARNING, args...)
}

func Err(args ...interface{}) {
	log.print(gsyslog.LOG_ERR, args...)
}

func Crit(args ...interface{}) {
	log.print(gsyslog.LOG_CRIT, args...)
}

func Alert(args ...interface{}) {
	log.print(gsyslog.LOG_ALERT, args...)
}

func Emerg(args ...interface{}) {
	log.print(gsyslog.LOG_EMERG, args...)
}

func Debugln(args ...interface{}) {
	log.println(gsyslog.LOG_DEBUG, args...)
}

func Infoln(args ...interface{}) {
	log.println(gsyslog.LOG_INFO, args...)
}

func Noticeln(args ...interface{}) {
	log.println(gsyslog.LOG_NOTICE, args...)
}

func Warningln(args ...interface{}) {
	log.print(gsyslog.LOG_WARNING, args...)
}

func Errln(args ...interface{}) {
	log.println(gsyslog.LOG_ERR, args...)
}

func Critln(args ...interface{}) {
	log.println(gsyslog.LOG_CRIT, args...)
}

func Alertln(args ...interface{}) {
	log.println(gsyslog.LOG_ALERT, args...)
}

func Emergln(args ...interface{}) {
	log.println(gsyslog.LOG_EMERG, args...)
}

func Debugf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_DEBUG, format, args...)
}

func Infof(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_INFO, format, args...)
}

func Noticef(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_NOTICE, format, args...)
}

func Warningf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_WARNING, format, args...)
}

func Errf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_ERR, format, args...)
}

func Critf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_CRIT, format, args...)
}

func Alertf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_ALERT, format, args...)
}

func Emergf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_EMERG, format, args...)
	os.Exit(1)
}

func Fatal(args ...interface{}) {
	log.println(gsyslog.LOG_ERR, args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	log.printf(gsyslog.LOG_ERR, format, args...)
	os.Exit(1)
}

// Spew will pretty print the args if the logger is set to DEBUG priority.
func Spew(args ...interface{}) {
	log.spew(args...)
}

func ChangePriority(level string) error {
	return log.adjustPriority(level)
}
