package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendTele(message []byte) []byte {
	tele_api := "https://api.telegram.org/bot{}/sendMessage"
	resp, err := http.Post(tele_api, "application/json", bytes.NewBuffer(message))
	if err != nil {
		fmt.Print(err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
