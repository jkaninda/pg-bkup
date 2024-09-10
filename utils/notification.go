package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

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
		Error("Message not sent, error: %s", string(body))
	}

}
func NotifySuccess(fileName string) {
	var vars = []string{
		"TG_TOKEN",
		"TG_CHAT_ID",
	}

	//Telegram notification
	err := CheckEnvVars(vars)
	if err == nil {
		message := "PostgreSQL Backup \n" +
			"Database has been backed up \n" +
			"Backup name is " + fileName
		sendMessage(message)
	}
}
func NotifyError(error string) {
	var vars = []string{
		"TG_TOKEN",
		"TG_CHAT_ID",
	}

	//Telegram notification
	err := CheckEnvVars(vars)
	if err == nil {
		message := "PostgreSQL Backup \n" +
			"An error occurred during database backup \n" +
			"Error: " + error
		sendMessage(message)
	}
}

func getTgUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", os.Getenv("TG_TOKEN"))

}
