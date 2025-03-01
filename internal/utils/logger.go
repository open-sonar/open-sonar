package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	infoLogger  = log.New(os.Stdout, "\033[1;34m[INFO]\033[0m ", log.Ldate|log.Ltime)
	warnLogger  = log.New(os.Stdout, "\033[1;33m[WARN]\033[0m ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "\033[1;31m[ERROR]\033[0m ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "\033[1;35m[DEBUG]\033[0m ", log.Ldate|log.Ltime)
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DebugLevel for verbose development logs
	DebugLevel LogLevel = iota
	// InfoLevel for general operational information
	InfoLevel
	// WarnLevel for warnings
	WarnLevel
	// ErrorLevel for errors
	ErrorLevel
)

var currentLevel LogLevel = InfoLevel

// SetLogLevel sets the current logging level
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// LogMessage logs a message at the specified level with caller info
func LogMessage(level LogLevel, msg string) {
	if level < currentLevel {
		return
	}

	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)
	logMsg := fmt.Sprintf("%s:%d %s", file, line, msg)

	switch level {
	case DebugLevel:
		debugLogger.Println(logMsg)
	case InfoLevel:
		infoLogger.Println(logMsg)
	case WarnLevel:
		warnLogger.Println(logMsg)
	case ErrorLevel:
		errorLogger.Println(logMsg)
	}
}

// Info logs an info level message
func Info(msg string) {
	LogMessage(InfoLevel, msg)
}

// Warn logs a warning level message
func Warn(msg string) {
	LogMessage(WarnLevel, msg)
}

// Error logs an error level message
func Error(msg string) {
	// Skip error logging in test mode if desired
	if os.Getenv("TEST_MODE") == "true" && os.Getenv("TEST_VERBOSE") != "true" {
		return
	}
	LogMessage(ErrorLevel, msg)
}

// Debug logs a debug level message
func Debug(msg string) {
	LogMessage(DebugLevel, msg)
}

// Timer provides a simple way to time operations
type Timer struct {
	Name      string
	StartTime time.Time
}

// NewTimer creates a new timer with the given name
func NewTimer(name string) *Timer {
	return &Timer{
		Name:      name,
		StartTime: time.Now(),
	}
}

// Stop stops the timer and logs the elapsed time
func (t *Timer) Stop() {
	elapsed := time.Since(t.StartTime)
	Info(fmt.Sprintf("%s took %s", t.Name, elapsed))
}
