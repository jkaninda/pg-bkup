// Package utils /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package utils

import (
	"fmt"
	"os"
	"time"
)

// Info message
func Info(msg string, args ...any) {
	var currentTime = time.Now().Format("2006/01/02 15:04:05")
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		fmt.Printf("%s INFO: %s\n", currentTime, msg)
	} else {
		fmt.Printf("%s INFO: %s\n", currentTime, formattedMessage)
	}
}

// Warn Warning message
func Warn(msg string, args ...any) {
	var currentTime = time.Now().Format("2006/01/02 15:04:05")
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		fmt.Printf("%s WARN: %s\n", currentTime, msg)
	} else {
		fmt.Printf("%s WARN: %s\n", currentTime, formattedMessage)
	}
}

// Error error message
func Error(msg string, args ...any) {
	var currentTime = time.Now().Format("2006/01/02 15:04:05")
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		fmt.Printf("%s ERROR: %s\n", currentTime, msg)
	} else {
		fmt.Printf("%s ERROR: %s\n", currentTime, formattedMessage)
	}
}
func Done(msg string, args ...any) {
	var currentTime = time.Now().Format("2006/01/02 15:04:05")
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		fmt.Printf("%s INFO: %s\n", currentTime, msg)
	} else {
		fmt.Printf("%s INFO: %s\n", currentTime, formattedMessage)
	}
}

func Fatal(msg string, args ...any) {
	var currentTime = time.Now().Format("2006/01/02 15:04:05")
	// Fatal logs an error message and exits the program.
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		fmt.Printf("%s ERROR: %s\n", currentTime, msg)
		NotifyError(msg)
	} else {
		fmt.Printf("%s ERROR: %s\n", currentTime, formattedMessage)
		NotifyError(formattedMessage)

	}

	os.Exit(1)
	os.Kill.Signal()
}
