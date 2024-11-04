// Package utils /
/*****
@author    Jonas Kaninda
@license   MIT License <https://opensource.org/licenses/MIT>
@Copyright © 2024 Jonas Kaninda
**/
package utils

import (
	"fmt"
	"log"
	"os"
)

// Info message
func Info(msg string, args ...any) {
	log.SetOutput(os.Stdout)
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		log.Printf("INFO: %s\n", msg)
	} else {
		log.Printf("INFO: %s\n", formattedMessage)
	}
}

// Warn a Warning message
func Warn(msg string, args ...any) {
	log.SetOutput(os.Stdout)
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		log.Printf("WARN: %s\n", msg)
	} else {
		log.Printf("WARN: %s\n", formattedMessage)
	}
}

// Error error message
func Error(msg string, args ...any) {
	log.SetOutput(os.Stdout)
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		log.Printf("ERROR: %s\n", msg)
	} else {
		log.Printf("ERROR: %s\n", formattedMessage)

	}
}
func Fatal(msg string, args ...any) {
	log.SetOutput(os.Stdout)
	// Fatal logs an error message and exits the program.
	formattedMessage := fmt.Sprintf(msg, args...)
	if len(args) == 0 {
		log.Printf("ERROR: %s\n", msg)
		NotifyError(msg)
	} else {
		log.Printf("ERROR: %s\n", formattedMessage)
		NotifyError(formattedMessage)

	}

	os.Exit(1)
}
