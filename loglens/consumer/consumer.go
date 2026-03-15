package main

import (
	"encoding/json"
	"fmt"
	"log"

	"loglens/models"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "my-first-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer:", err)
	}
	defer consumer.Close()

	consumer.SubscribeTopics([]string{"raw-logs"}, nil)

	for {
		// ReadMessage reads a single message from the consumer. The timeout parameter specifies how long to wait for a message.
		// Passing -1 as the timeout means the call will block indefinitely until a message is received or an error occurs.
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Println("Error reading message:", err)
			continue
		}
		var event models.LogEvent
		err = json.Unmarshal(msg.Value, &event) // Deserializes json and stores in the event variable of type logevents
		if err != nil {
			log.Println("Failed to parse event ", err)
			continue
		}

		if event.Level == "info" {
			continue
		}

		// fmt.Println("Received:", event)
		// fmt.Println("ID = ", event.ID)
		fmt.Println("Service = ", event.Service)
		// fmt.Println("Level = ", event.Level)
		// fmt.Println("Message = ", event.Message)
		// fmt.Println("Timestamp = ", event.Timestamp)
		fmt.Println("Meta = ", event.Meta["host"], event.Meta["env"])
	}
}
