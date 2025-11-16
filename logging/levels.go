package logging

// A Level represents a logging level.
type Level uint8

// The following constants represent logging levels in increasing levels of seriousness.
const (
	// LevelDebug are debug output useful during program testing
	// and debugging.
	LevelDebug = 1 << iota

	// LevelInfo is used for informational messages.
	LevelInfo

	// LevelWarning is for messages that are warning conditions:
	// they're not indicative of a failure, but of a situation
	// that may lead to a failure later.
	LevelWarning

	// LevelError is for messages indicating an error of some
	// kind.
	LevelError

	// LevelCritical are messages for critical conditions.
	LevelCritical

	// LevelFatal messages are akin to syslog's LOG_EMERG: the
	// system is unusable and cannot continue execution.
	LevelFatal
)

// DefaultLevel is the default logging level when none is provided.
const DefaultLevel = LevelInfo

var levelPrefix = [...]string{
	LevelDebug:    "DEBUG",
	LevelInfo:     "INFO",
	LevelWarning:  "WARNING",
	LevelError:    "ERROR",
	LevelCritical: "CRITICAL",
	LevelFatal:    "FATAL",
}

// DateFormat contains the default date format string used by the logger.
const DateFormat = "2006-01-02T15:03:04-0700"
