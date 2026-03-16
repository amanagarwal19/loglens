package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"loglens/cluster"
	"loglens/embeddings"
	"loglens/models"
	"net/http"
	"os"
	"sort"

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
	redisStore := cluster.NewStore("localhost:6379")

	// Create the main engine.
	engine := cluster.New(embedSvc, 0.8, redisStore)

	go startHTTPServer(engine)

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

// HTTP Server to expose current clusters for visualization. In a real app, this would be a separate service.
func startHTTPServer(engine *cluster.Engine) {
	server := http.NewServeMux()

	server.HandleFunc("GET /api/clusters/all", func(w http.ResponseWriter, r *http.Request) {

		clusters := engine.All()

		// sort by last seen so most recent errors appear first
		sort.Slice(clusters, func(i, j int) bool {
			return clusters[i].LastSeen.After(clusters[j].LastSeen)
		})

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(clusters)
	})

	server.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := engine.Reset(context.Background()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "clusters reset")
	})

	log.Println("consumer API running at http://localhost:4000")
	http.ListenAndServe(":4000", server)

}
