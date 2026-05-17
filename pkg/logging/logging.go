// Package logging provides structured logging for the agent SDK.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging.
type Logger struct {
	mu    sync.Mutex
	level Level
	log   *log.Logger
}

// New creates a new Logger that writes to stderr at INFO level.
func New() *Logger {
	return NewWithWriter(os.Stderr, LevelInfo)
}

// NewWithWriter creates a Logger with a custom writer and level.
func NewWithWriter(w io.Writer, level Level) *Logger {
	return &Logger{
		level: level,
		log:   log.New(w, "", log.LstdFlags),
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) logf(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level < l.level {
		return
	}
	l.log.Printf("[%s] %s", level.String(), fmt.Sprintf(format, args...))
}

func (l *Logger) Debug(format string, args ...interface{}) { l.logf(LevelDebug, format, args...) }
func (l *Logger) Info(format string, args ...interface{})  { l.logf(LevelInfo, format, args...) }
func (l *Logger) Warn(format string, args ...interface{})  { l.logf(LevelWarn, format, args...) }
func (l *Logger) Error(format string, args ...interface{}) { l.logf(LevelError, format, args...) }
