package main

import (
	"encoding/json"
	"fmt"

	bouncemailcounter "github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter"
	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/client"
	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/utils"
	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
)

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func Task(duration int) {
	cfg := bouncemailcounter.Load()
	message := `
		Counter_Time: %v
		per_minute: %d
	`
	time_from, counter := utils.Counter(duration)
	message = fmt.Sprintf(message, time_from, counter)
	logrus.Info(message)
	req_body := &sendMessageReqBody{
		ChatID: cfg.TelegramChatId,
		Text:   message,
	}
	req_bytes, err := json.Marshal(req_body)
	if err != nil {
		logrus.Error(err)
	}
	if counter > 10 {
		logrus.Info("Sending to telegram.....")
		client.SendTele(req_bytes)
		logrus.Info("Done")
	}
}

func main() {
	utils.InitLog()
	gocron.Every(1).Second().Do(Task, 1)
	<-gocron.Start()
}
