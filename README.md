# bounce-mail-counter

## Requied environments

```sh
export KAFKA_TOPIC=""
export CONSUMER_GROUP=""
export KAFKA_BROKER=""
export MONGO_URI=""
export MONGO_DATABASE=""
export MONGO_COLLECTION=""
export TELEGRAM_CHAT_ID=
export TELEGRAM_BOT_TOKEN=""
```

# Run:

```sh 
cd bounce-mail-counter
```

- ### App

	```sh
	go run bounce_mail_counter/app/app.go
	```

- ### Worker

	```sh
	go run bounce_mail_counter/worker/kafka_worker.go
	```