package log

import (
	"fmt"
	"os"
)

// Logger is a logger with debug capabilities
type Logger interface {
	// Print prints an info message
	Print(...interface{})
	// Printf prints a formatted info message
	Printf(string, ...interface{})
	// Debug prints a debug message (if verbose set)
	Debug(...interface{})
	// Debugf prints a formatted debug message (if verbose set)
	Debugf(string, ...interface{})
	// SetVerbose changes the behavior of the logger in regards to debugging
	// messages
	SetVerbose(bool)
	// GetVerbose returns the verbosity
	GetVerbose() bool
}

var logger Logger = new(defaultLogger)

func Print(a ...interface{})                 { logger.Print(a...) }
func Printf(format string, a ...interface{}) { logger.Printf(format, a...) }
func Debug(a ...interface{})                 { logger.Debug(a...) }
func Debugf(format string, a ...interface{}) { logger.Debugf(format, a...) }

func SetLogger(l Logger) {
	logger = l
}

// defaultLogger writes messages to stdout
type defaultLogger struct {
	verbose bool
}

func (_ *defaultLogger) Print(a ...interface{}) {
	fmt.Fprint(os.Stdout, a...)
}

func (_ *defaultLogger) Printf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func (l *defaultLogger) Debug(a ...interface{}) {
	if !l.verbose {
		return
	}
	fmt.Fprint(os.Stdout, a...)
}

func (l *defaultLogger) Debugf(format string, a ...interface{}) {
	if !l.verbose {
		return
	}
	fmt.Fprintf(os.Stdout, format, a...)
}

func (l *defaultLogger) SetVerbose(v bool) {
	l.verbose = v
}

func (l *defaultLogger) GetVerbose() bool {
	return l.verbose
}
