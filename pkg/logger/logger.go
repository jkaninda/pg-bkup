package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

// Info returns info log
func Info(msg string, args ...interface{}) {
	log.SetOutput(getStd("/dev/stdout"))
	logWithCaller("INFO", msg, args...)

}

// Warn returns warning log
func Warn(msg string, args ...interface{}) {
	log.SetOutput(getStd("/dev/stdout"))
	logWithCaller("WARN", msg, args...)

}

// Error logs error messages
func Error(msg string, args ...interface{}) {
	log.SetOutput(getStd("/dev/stderr"))
	logWithCaller("ERROR", msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	log.SetOutput(os.Stdout)
	logWithCaller("ERROR", msg, args...)
	os.Exit(1)
}

// Helper function to format and log messages with file and line number
func logWithCaller(level, msg string, args ...interface{}) {
	// Format message if there are additional arguments
	formattedMessage := msg
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(msg, args...)
	}

	// Get the caller's file and line number (skip 2 frames)
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	// Log message with caller information if GOMA_LOG_LEVEL is trace
	if strings.ToLower(level) != "off" {
		if strings.ToLower(level) == traceLog {
			log.Printf("%s: %s (File: %s, Line: %d)\n", level, formattedMessage, file, line)
		} else {
			log.Printf("%s: %s\n", level, formattedMessage)
		}
	}
}

func getStd(out string) *os.File {
	switch out {
	case "/dev/stdout":
		return os.Stdout
	case "/dev/stderr":
		return os.Stderr
	case "/dev/stdin":
		return os.Stdin
	default:
		return os.Stdout

	}
}
