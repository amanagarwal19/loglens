package main

import (
	"encoding/json"
	"fmt"
	"log"
	"loglens/models"
	"math/rand"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

var messages = []string{
	"database connection refused: port 5432",
	"database connection timeout after 30s",
	"nil pointer dereference in handler /api/users",
	"nil pointer dereference in handler /api/orders",
	"failed to parse request body: unexpected EOF",
	"failed to unmarshal JSON: unexpected EOF",
	"redis: connection pool exhausted",
	"redis: connection refused",
	"context deadline exceeded calling payments service",
	"context deadline exceeded calling inventory service",
}

var services = []string{
	"api-gateway",
	"user-service",
	"order-service",
	"payment-service",
}

var levels = []string{
	"error",
	"warning",
	"info",
}

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer producer.Close()
	topic := "raw-logs"

	for {
		event := models.LogEvent{
			ID:        uuid.New().String(),
			Service:   services[rand.Intn(len(services))],
			Level:     levels[rand.Intn(len(levels))],
			Message:   messages[rand.Intn(len(messages))],
			Timestamp: time.Now(),
			Meta: map[string]string{
				"host": fmt.Sprintf("server-%d", rand.Intn(10)+1),
				"env":  "production",
			},
		}

		data, err := json.Marshal(event) // Returns byte[] and error
		if err != nil {
			log.Fatal("Failed to marshal event:", err)
		}

		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: data,
			Key:   []byte(event.Service),
		}, nil)

		// producer.Flush(2000)
		fmt.Printf("Message sent:[%s] - %s\n - %s", event.Service, event.Meta["host"], event.Meta["env"])
		time.Sleep(5000 * time.Millisecond)
	}

}
