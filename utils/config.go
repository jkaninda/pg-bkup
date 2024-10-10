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
	BackupSize      int64
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
