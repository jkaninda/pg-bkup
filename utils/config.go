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
	File           string
	BackupSize     int64
	Database       string
	StartTime      string
	EndTime        string
	Storage        string
	BackupLocation string
}
type ErrorMessage struct {
	Database string
	EndTime  string
	Error    string
}

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

const templatePath = "/config/templates"
