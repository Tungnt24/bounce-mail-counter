package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client_instance *mongo.Client

//Used during creation of singleton client object in GetMongoClient().
var client_instance_error error

//Used to execute client creation procedure only once.
var mongoOnce sync.Once

//I have used below constants just to hold required database config's.
const (
	CONNECTIONSTRING = ""
	DB               = ""
	LOGS             = ""
)

type Log struct {
	Queue_Id              string `bson:"queue_id"`
	From                  string `bson:"from"`
	To                    string `bson:"to"`
	Message_Id            string `bson:"message_id"`
	Recipient_Smtp_Ip     string `bson:"recipient_smtp_ip"`
	Recipient_Smtp_Domain string `bson:"recipient_smtp_domain"`
	Status                string `bson:"status"`
	Message               string `bson:"message"`
	Sent_At               string `bson:"sent_at"`
}

func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		client_options := options.Client().ApplyURI(CONNECTIONSTRING)
		client, err := mongo.Connect(context.TODO(), client_options)
		if err != nil {
			client_instance_error = err
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			client_instance_error = err
		}
		client_instance = client
	})
	return client_instance, client_instance_error
}
func CreateLog(task Log) error {
	client, err := GetMongoClient()
	if err != nil {
		return err
	}
	collection := client.Database(DB).Collection(LOGS)
	_, err = collection.InsertOne(context.TODO(), task)
	if err != nil {
		return err
	}
	return nil
}

func UpdateLog(queue_id string, key string, value string) error {
	filter := bson.D{primitive.E{Key: "queue_id", Value: queue_id}}

	updater := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: key, Value: value},
	}}}

	client, err := GetMongoClient()
	if err != nil {
		return err
	}
	collection := client.Database(DB).Collection(LOGS)

	_, err = collection.UpdateOne(context.TODO(), filter, updater)
	if err != nil {
		return err
	}
	return nil
}

func GetLogByQueueId(queue_id string) (Log, error) {
	result := Log{}
	//Define filter query for fetching specific document from collection
	filter := bson.D{primitive.E{Key: "queue_id", Value: queue_id}}
	//Get MongoDB connection using connectionhelper.
	client, err := GetMongoClient()
	if err != nil {
		return result, err
	}
	//Create a handle to the respective collection in the database.
	collection := client.Database(DB).Collection(LOGS)
	//Perform FindOne operation & validate against the error.
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	//Return result without any error.
	return result, nil
}

func FilterMessage(message string) bool {
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

func CollectField(raw_message_str string) (Log, error) {
	log := Log{}
	mapping := Dump(raw_message_str)
	raw_message := fmt.Sprintf("%v\n", mapping["message"])
	fmt.Println("\nMESSAGE: ", raw_message)
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	fmt.Fprintln(f, raw_message)
	if err != nil {
		panic(err)
	}
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
			log.From = value
		case "To":
			log.To = value
			log.Sent_At = strings.Trim(timestamp, "\n")
		case "Message-Id":
			log.Message_Id = value
		case "Relay":
			open_char := strings.Index(value, "[")
			close_char := strings.Index(value, "]")
			log.Recipient_Smtp_Domain = value[:open_char]
			log.Recipient_Smtp_Ip = value[open_char+1 : close_char]
		case "Status":
			status_message := raw_message[status_message_index:]
			log.Status = value
			log.Message = status_message
		}
	}
	log.Queue_Id = queue_id
	return log, nil
}

func AggregateLog(log Log) {
	v := reflect.ValueOf(log)
	typeOfS := v.Type()
	result, _ := GetLogByQueueId(log.Queue_Id)
	f, _ := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	r := fmt.Sprintf("\nRESULT: %s", result)
	fmt.Fprintln(f, r)
	if result != (Log{}) {
		for i := 1; i < v.NumField(); i++ {
			key := strings.ToLower(typeOfS.Field(i).Name)
			value := fmt.Sprintf("%v", v.Field(i).Interface())
			if value == "" {
				continue
			}
			a := fmt.Sprintf("\n%s : KEY: %s | VALUE: %s", log.Queue_Id, key, value)
			fmt.Fprintln(f, a)
			UpdateLog(log.Queue_Id, key, value)
		}
	} else {
		CreateLog(log)
	}
}

func DetectSpam(message string) bool {
	spam_pattern := `^.*(\bstatus\=bounced\b).*(\breputation\b|\bspam\b|\bspamhaus\b|\blisted\b|\bblock\b|\bblocked\b|\bsecurity\b|\bblacklisted\b|\bphish\b|\bphishing\b|\bvirus\b|\brejected\b|\bblacklisted\b|\bblacklist\b).*($|[^\w])`
	match, _ := regexp.MatchString(spam_pattern, message)
	return match
}
