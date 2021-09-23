package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	bouncemailcounter "github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter"
)

func SendTele(message []byte) []byte {
	cfg := bouncemailcounter.Load()
	tele_api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TelegramBotToken)
	resp, err := http.Post(tele_api, "application/json", bytes.NewBuffer(message))
	if err != nil {
		fmt.Print(err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
