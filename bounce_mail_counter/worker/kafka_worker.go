package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/Tungnt24/bounce-mail-counter/bounce_mail_counter/utils"
)

type Consumer struct {
	ready chan bool
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for raw_message := range claim.Messages() {
		raw_message_str := string(raw_message.Value)
		is_legal_message := utils.FilterLog(raw_message_str)
		if is_legal_message {
			log, _ := utils.CollectField(raw_message_str)
			utils.AggregateLog(log)
		}

		session.MarkMessage(raw_message, "")
	}
	return nil
}

func ConnectConsumer(brokersUrl []string, groupId string) (sarama.ConsumerGroup, error) {
	log.Println("Starting a new Sarama consumer")
	sarama.Logger = log.New(os.Stdout, "[mail_counter] ", log.LstdFlags)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	client, err := sarama.NewConsumerGroup(brokersUrl, groupId, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}
	return client, nil
}

func Worker() {
	consumer := Consumer{
		ready: make(chan bool),
	}
	ctx, cancel := context.WithCancel(context.Background())
	broker := []string{""}
	topics := ""
	group_id := ""
	client, err := ConnectConsumer(broker, group_id)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

func main() {
	Worker()
}
