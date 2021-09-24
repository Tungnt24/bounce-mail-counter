package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/client"
	"github.com/sirupsen/logrus"
)

func FilterLog(message string) bool {
	if strings.Contains(message, "capt-se") {
		if strings.Contains(message, "postfix/qmgr") &&
			strings.Contains(message, "from=") ||
			strings.Contains(message, "postfix/cleanup") &&
				strings.Contains(message, "message-id") ||
			strings.Contains(message, "postfix/smtp") &&
				strings.Contains(message, "status=bounced") {
			return true
		}
	}

	return false
}

func Dump(message string) map[string]interface{} {
	mapping := make(map[string]interface{})
	err := json.Unmarshal([]byte(message), &mapping)

	if err != nil {
		panic(err)
	}
	return mapping
}

func ConvertToTimeUTC(time_str string) time.Time {
	layout := "2006-01-02T15:04:05Z"
	time_parse, _ := time.Parse(layout, time_str)
	return time_parse
}

func CollectField(raw_message_str string) (client.MailLog, error) {
	mail_log := client.MailLog{}
	mapping := Dump(raw_message_str)
	raw_message := fmt.Sprintf("%v\n", mapping["message"])
	logrus.Info("Message: ", raw_message)
	timestamp := fmt.Sprintf("%v\n", mapping["@timestamp"])
	index := strings.Index(raw_message, "]:")
	status_message_index := strings.Index(raw_message, "(")
	if status_message_index == -1 {
		status_message_index = len(raw_message) - 1
	}
	message := strings.TrimSpace(raw_message[index+2 : status_message_index])
	fields := strings.Split(message, " ")
	re := regexp.MustCompile(`=`)
	queue_id := strings.Trim(strings.Replace(string(fields[0]), ":", "", 1), "")
	for _, field := range fields[1:] {
		items := re.Split(field, 2)
		if len(items) <= 1 {
			continue
		}
		raw_key, raw_value := strings.Trim(items[0], " "), strings.Trim(items[1], " ")
		key := strings.Title(raw_key)
		value := strings.Replace(raw_value, ",", " ", 1)
		switch key {
		case "From":
			mail_log.From = value
		case "To":
			mail_log.To = value
		case "Message-Id":
			mail_log.Message_Id = value
		case "Relay":
			open_char := strings.Index(value, "[")
			close_char := strings.Index(value, "]")
			if open_char == -1 || close_char == -1 {
				continue
			}
			mail_log.Recipient_Smtp_Domain = value[:open_char]
			mail_log.Recipient_Smtp_Ip = value[open_char+1 : close_char]
		case "Status":
			mail_log.Sent_At = ConvertToTimeUTC(strings.Trim(timestamp, "\n"))
			status_message := raw_message[status_message_index:]
			mail_log.Status = value
			mail_log.Message = status_message
		}
	}
	mail_log.Queue_Id = queue_id
	return mail_log, nil
}

func AggregateLog(mail_log client.MailLog) {
	v := reflect.ValueOf(mail_log)
	typeOfS := v.Type()
	result, _ := client.GetLogByQueueId(mail_log.Queue_Id)
	if result != (client.MailLog{}) {
		for i := 1; i < v.NumField(); i++ {
			key := strings.ToLower(typeOfS.Field(i).Name)
			value := fmt.Sprintf("%v", v.Field(i).Interface())
			if value == "" || value == "0001-01-01 00:00:00 +0000 UTC" {
				continue
			}
			client.UpdateLog(mail_log.Queue_Id, key, value)
		}
	} else {
		client.CreateLog(mail_log)
	}
}

func DetectSpam(message string) bool {
	spam_pattern := `^.*(\breputation\b|\bspam\b|\bspamhaus\b|\blisted\b|\bblock\b|\bblocked\b|\bsecurity\b|\bblacklisted\b|\bphish\b|\bphishing\b|\bvirus\b|\brejected\b|\bblacklisted\b|\bblacklist\b).*($|[^\w])`
	message = strings.ToLower(message)
	match, _ := regexp.MatchString(spam_pattern, message)
	return match
}

func GetBounceMail(from_datetime time.Time, to_datetime time.Time) []client.MailLog {
	key := "status"
	value := "bounced"
	result, _ := client.GetManyLogs(key, value, from_datetime, to_datetime)
	return result
}

func GetTime(duration int) (time.Time, time.Time) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)
	now_utc_str := now.Format(time.RFC3339)
	then_str := now.Add(time.Duration(-duration) * time.Minute).In(loc).Format(time.RFC3339)
	to_datetime := ConvertToTimeUTC(now_utc_str)
	from_datetime := ConvertToTimeUTC(then_str)
	return from_datetime, to_datetime
}

func Counter(duration int) (time.Time, int) {
	from, to := GetTime(duration)
	counter := 0
	result := GetBounceMail(from, to)
	for _, value := range result {
		is_spam := DetectSpam(value.Message)
		if is_spam {
			counter += 1
		}
	}
	return from, counter
}
