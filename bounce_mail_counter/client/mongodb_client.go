package client

import (
	"context"
	"sync"
	"time"

	bouncemailcounter "github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client_instance *mongo.Client
var client_instance_error error
var mongoOnce sync.Once

var (
	CONNECTIONSTRING = bouncemailcounter.Load().MongoUri
	DB               = bouncemailcounter.Load().MongoDataBase
	MailLogs         = bouncemailcounter.Load().MongoCollection
)

type MailLog struct {
	Queue_Id              string    `bson:"queue_id"`
	From                  string    `bson:"from"`
	To                    string    `bson:"to"`
	Message_Id            string    `bson:"message_id"`
	Recipient_Smtp_Ip     string    `bson:"recipient_smtp_ip"`
	Recipient_Smtp_Domain string    `bson:"recipient_smtp_domain"`
	Status                string    `bson:"status"`
	Message               string    `bson:"message"`
	Sent_At               time.Time `bson:"sent_at"`
}

func ConvertToTimeMST(time_str string) time.Time {
	layout := "2006-01-02 15:04:05 -0700 MST"
	time_parse, _ := time.Parse(layout, time_str)
	return time_parse
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
func CreateLog(task MailLog) error {
	client, err := GetMongoClient()
	if err != nil {
		return err
	}
	collection := client.Database(DB).Collection(MailLogs)
	_, err = collection.InsertOne(context.TODO(), task)
	if err != nil {
		return err
	}
	return nil
}

func GetManyLogs(key string, value string, from_date time.Time, to_date time.Time) ([]MailLog, error) {
	filter := bson.M{
		"sent_at": bson.M{
			"$gte": primitive.NewDateTimeFromTime(from_date),
			"$lt":  primitive.NewDateTimeFromTime(to_date),
		},
		key: value,
	}
	mail_log := []MailLog{}

	client, err := GetMongoClient()
	if err != nil {
		return mail_log, err
	}
	collection := client.Database(DB).Collection(MailLogs)

	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return mail_log, err
	}

	for cur.Next(context.TODO()) {
		t := MailLog{}
		err := cur.Decode(&t)
		if err != nil {
			return mail_log, err
		}
		mail_log = append(mail_log, t)
	}

	cur.Close(context.TODO())
	if len(mail_log) == 0 {
		return mail_log, mongo.ErrNoDocuments
	}
	return mail_log, nil
}

func UpdateLog(queue_id string, key string, value string) error {
	filter := bson.D{primitive.E{Key: "queue_id", Value: queue_id}}
	updater := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: key, Value: value},
	}}}
	if key == "sent_at" {
		updater = bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: key, Value: ConvertToTimeMST(value)},
		}}}
	}

	client, err := GetMongoClient()
	if err != nil {
		return err
	}
	collection := client.Database(DB).Collection(MailLogs)

	_, err = collection.UpdateOne(context.TODO(), filter, updater)
	if err != nil {
		return err
	}
	return nil
}

func GetLogByQueueId(queue_id string) (MailLog, error) {
	result := MailLog{}
	filter := bson.D{primitive.E{Key: "queue_id", Value: queue_id}}
	client, err := GetMongoClient()
	if err != nil {
		return result, err
	}
	collection := client.Database(DB).Collection(MailLogs)
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}
