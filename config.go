package bouncemailcounter

type Config struct {
	KafkaBroker        []string `envconfig:"KAFKA_BROKER"`
	KafkaTopic         string   `envconfig:"KAFKA_TOPIC"`
	KafkaConsumerGroup string   `envconfig:"CONSUMER_GROUP"`

	MongoUri        string `envconfig:"MONGO_URI"`
	MongoDataBase   string `envconfig:"MONGO_DATABASE"`
	MongoCollection string `envconfig:"MONGO_COLLECTION"`

	TelegramChatId   int64  `envconfig:"TELEGRAM_CHAT_ID"`
	TelegramBotToken string `envconfig:"TELEGRAM_BOT_TOKEN"`
}
