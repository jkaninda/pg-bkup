/*
MIT License

Copyright (c) 2023 Jonas Kaninda

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package utils

import "os"

type MailConfig struct {
	MailHost     string
	MailPort     int
	MailUserName string
	MailPassword string
	MailTo       string
	MailFrom     string
	SkipTls      bool
}
type NotificationData struct {
	File            string
	BackupSize      string
	Database        string
	StartTime       string
	EndTime         string
	Storage         string
	BackupLocation  string
	BackupReference string
}
type ErrorMessage struct {
	Database        string
	EndTime         string
	Error           string
	BackupReference string
	DatabaseName    string
}

// loadMailConfig gets mail environment variables and returns MailConfig
func loadMailConfig() *MailConfig {
	return &MailConfig{
		MailHost:     os.Getenv("MAIL_HOST"),
		MailPort:     GetIntEnv("MAIL_PORT"),
		MailUserName: os.Getenv("MAIL_USERNAME"),
		MailPassword: os.Getenv("MAIL_PASSWORD"),
		MailTo:       os.Getenv("MAIL_TO"),
		MailFrom:     os.Getenv("MAIL_FROM"),
		SkipTls:      os.Getenv("MAIL_SKIP_TLS") == "false",
	}

}

// TimeFormat returns the format of the time
func TimeFormat() string {
	format := os.Getenv("TIME_FORMAT")
	if format == "" {
		return "2006-01-02 at 15:04:05"

	}
	return format
}

func backupReference() string {
	return os.Getenv("BACKUP_REFERENCE")
}

const templatePath = "/config/templates"

var DatabaseName = ""
