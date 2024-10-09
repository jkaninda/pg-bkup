package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/go-mail/mail"
	"github.com/robfig/cron/v3"
	"html/template"
	"io/ioutil"
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

func SendEmail(subject, body string) {
	Info("Start sending email....")
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
		Fatal("Error could not send email : %v", err)
	}
	Info("Email has been sent")

}
func sendMessage(msg string) {

	Info("Sending notification... ")
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
		panic(err)
	}
	code := response.StatusCode
	if code == 200 {
		Info("Notification has been sent")
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		Fatal("Error could not send message, error: %s", string(body))
	}

}
func NotifySuccess(notificationData *NotificationData) {
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

	//Email notification
	err := CheckEnvVars(mailVars)
	if err == nil {
		body, err := parseTemplate(*notificationData, "email.template")
		if err != nil {
			Error("Could not parse email template: %v", err)
		}
		SendEmail(fmt.Sprintf("✅  Database Backup Notification – %s", notificationData.Database), body)
	}
	//Telegram notification
	err = CheckEnvVars(vars)
	if err == nil {
		message, err := parseTemplate(*notificationData, "telegram.template")
		if err != nil {
			Error("Could not parse email template: %v", err)
		}

		sendMessage(message)
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

	//Email notification
	err := CheckEnvVars(mailVars)
	if err == nil {
		body, err := parseTemplate(ErrorMessage{
			Error:   error,
			EndTime: time.Now().Format("2006-01-02 15:04:05"),
		}, "email-error.template")
		if err != nil {
			Error("Could not parse email template: %v", err)
		}
		SendEmail(fmt.Sprintf("🔴 Urgent: Database Backup Failure Notification"), body)
	}
	//Telegram notification
	err = CheckEnvVars(vars)
	if err == nil {
		message, err := parseTemplate(ErrorMessage{
			Error:   error,
			EndTime: time.Now().Format("2006-01-02 15:04:05"),
		}, "telegram-error.template")
		if err != nil {
			Error("Could not parse email template: %v", err)
		}

		sendMessage(message)
	}
}

func getTgUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", os.Getenv("TG_TOKEN"))

}
func IsValidCronExpression(cronExpr string) bool {
	_, err := cron.ParseStandard(cronExpr)
	return err == nil
}
