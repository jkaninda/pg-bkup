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

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-mail/mail"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func parseTemplate[T any](data T, fileName string) (string, error) {
	// Open the file
	tmpl, err := template.ParseFiles(filepath.Join(templatePath, fileName))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func SendEmail(subject, body string) error {
	Info("Start sending email notification....")
	config := loadMailConfig()
	emails := strings.Split(config.MailTo, ",")
	m := mail.NewMessage()
	m.SetHeader("From", config.MailFrom)
	m.SetHeader("To", emails...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := mail.NewDialer(config.MailHost, config.MailPort, config.MailUserName, config.MailPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: config.SkipTls}

	if err := d.DialAndSend(m); err != nil {
		Error("Error could not send email : %v", err)
		return err
	}
	Info("Email notification has been sent")
	return nil

}
func sendMessage(msg string) error {

	Info("Sending Telegram notification... ")
	chatId := os.Getenv("TG_CHAT_ID")
	body, _ := json.Marshal(map[string]string{
		"chat_id": chatId,
		"text":    msg,
	})
	url := fmt.Sprintf("%s/sendMessage", getTgUrl())
	// Create an HTTP post request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	code := response.StatusCode
	if code == 200 {
		Info("Telegram notification has been sent")
		return nil
	} else {
		body, _ := io.ReadAll(response.Body)
		Error("Error could not send message, error: %s", string(body))
		return fmt.Errorf("error could not send message %s", string(body))
	}

}
func NotifySuccess(notificationData *NotificationData) {
	notificationData.BackupReference = backupReference()
	var vars = []string{
		"TG_TOKEN",
		"TG_CHAT_ID",
	}
	var mailVars = []string{
		"MAIL_HOST",
		"MAIL_PORT",
		"MAIL_USERNAME",
		"MAIL_PASSWORD",
		"MAIL_FROM",
		"MAIL_TO",
	}

	// Email notification
	err := CheckEnvVars(mailVars)
	if err == nil {
		body, err := parseTemplate(*notificationData, "email.tmpl")
		if err != nil {
			Error("Could not parse email template: %v", err)
		}
		err = SendEmail(fmt.Sprintf("✅  Database Backup Notification – %s", notificationData.Database), body)
		if err != nil {
			Error("Could not send email: %v", err)
		}
	}
	// Telegram notification
	err = CheckEnvVars(vars)
	if err == nil {
		message, err := parseTemplate(*notificationData, "telegram.tmpl")
		if err != nil {
			Error("Could not parse telegram template: %v", err)
		}

		err = sendMessage(message)
		if err != nil {
			Error("Could not send Telegram message: %v", err)
		}
	}
}
func NotifyError(error string) {
	var vars = []string{
		"TG_TOKEN",
		"TG_CHAT_ID",
	}
	var mailVars = []string{
		"MAIL_HOST",
		"MAIL_PORT",
		"MAIL_USERNAME",
		"MAIL_PASSWORD",
		"MAIL_FROM",
		"MAIL_TO",
	}

	// Email notification
	err := CheckEnvVars(mailVars)
	if err == nil {
		body, err := parseTemplate(ErrorMessage{
			Error:           error,
			EndTime:         time.Now().Format(TimeFormat()),
			BackupReference: os.Getenv("BACKUP_REFERENCE"),
			DatabaseName:    DatabaseName,
		}, "email-error.tmpl")
		if err != nil {
			Error("Could not parse error template: %v", err)
		}
		err = SendEmail("🔴 Urgent: Database Backup Failure Notification", body)
		if err != nil {
			Error("Could not send email: %v", err)
		}
	}
	// Telegram notification
	err = CheckEnvVars(vars)
	if err == nil {
		message, err := parseTemplate(ErrorMessage{
			Error:           error,
			EndTime:         time.Now().Format(TimeFormat()),
			BackupReference: os.Getenv("BACKUP_REFERENCE"),
			DatabaseName:    DatabaseName,
		}, "telegram-error.tmpl")
		if err != nil {
			Error("Could not parse error template: %v", err)

		}

		err = sendMessage(message)
		if err != nil {
			Error("Could not send telegram message: %v", err)
		}
	}
}

func getTgUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", os.Getenv("TG_TOKEN"))

}
