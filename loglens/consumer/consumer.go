package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"loglens/cluster"
	"loglens/models"
	"os"

	"loglens/embeddings"

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

	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey == "" {
		log.Fatal("API KEY NOT FOUND")
	}

	consumer.SubscribeTopics([]string{"raw-logs"}, nil)
	embedSvc := embeddings.New(apiKey)

	engine := cluster.New(embedSvc, 0.8)

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

		cluster, isNew, err := engine.Add(context.Background(), event)

		if err != nil {
			fmt.Printf("Error adding to cluster %s", err)
			continue
		}
		if isNew {
			fmt.Printf("NEW CLUSTER [%s] %s\n", event.Service, event.Message)
		} else {
			fmt.Printf("merged into cluster %s (count=%d)\n", cluster.ID, cluster.Count)
		}
	}
}
