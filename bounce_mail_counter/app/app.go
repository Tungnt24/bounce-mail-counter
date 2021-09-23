package main

import (
	"encoding/json"
	"fmt"

	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/client"
	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/utils"
	"github.com/jasonlvhit/gocron"
)

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func Task(duration int) {
	message := `
		Counter_Time: %v
		per_minute: %d
	`
	time_from, counter := utils.Counter(duration)
	message = fmt.Sprintf(message, time_from, counter)
	req_body := &sendMessageReqBody{
		ChatID: 932131897,
		Text:   message,
	}
	req_bytes, err := json.Marshal(req_body)
	if err != nil {
		fmt.Print(err)
	}
	if counter > 10 {
		client.SendTele(req_bytes)
	}
}

func main() {
	gocron.Every(1).Minute().Do(Task, 1)
	//gocron.Every(1).Hour().Do(Task, 60)
	//gocron.Every(1).Day().Do(Task, 24)
	<-gocron.Start()
}
