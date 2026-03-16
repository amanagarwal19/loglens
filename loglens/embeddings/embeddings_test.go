package embeddings

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// func TestConnection(t *testing.T) {
// 	apiKey := os.Getenv("VOYAGE_API_KEY")
// 	if apiKey == "" {
// 		fmt.Println("VOYAGE_API_KEY not set, skipping test")
// 		t.Skip("VOYAGE_API_KEY not set, skipping test")
// 	}

// 	svc := New(apiKey)
// 	msg, err := svc.Embed(context.Background(), "test connection")
// 	if err != nil {
// 		fmt.Println("failed to connect to Voyage API:", err)
// 		t.Fatal("failed to connect to Voyage API:", err)
// 	}
// 	fmt.Println("successfully connected to Voyage API")
// 	fmt.Println("voyage API Response message:", msg)
// }

func TestEmbed(t *testing.T) {
	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey == "" {
		t.Skip("VOYAGE_API_KEY not set, skipping test")
	}

	svc := New(apiKey)

	vec, err := svc.Embed(context.Background(), "database connection refused")
	if err != nil {
		t.Fatal("embed failed:", err)
	}

	fmt.Println("vector length:", len(vec))
	fmt.Println("first 5 values:", vec)
}

// func TestCosineSimilarity() {
// 	svc := New("")

// 	vec1, _ := svc.Embed(context.Background(), "database connection refused port 5432")
// 	vec2, _ := svc.Embed(context.Background(), "database connection refused port 5433")
// 	vec3, _ := svc.Embed(context.Background(), "nil pointer dereference in handler")

// 	sim12 := CosineSimilarity(vec1, vec2)
// 	sim13 := CosineSimilarity(vec1, vec3)

// 	fmt.Printf("similarity between similar errors: %.4f\n", sim12)
// 	fmt.Printf("similarity between different errors: %.4f\n", sim13)

// 	if sim12 <= sim13 {
// 		fmt.Println("expected similar errors to have higher similarity than different errors")
// 	}
// }
