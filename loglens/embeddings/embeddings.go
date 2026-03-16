package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

type Vector []float32

type request struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type response struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

type Service struct {
	apiKey string
	client *http.Client
}

func New(apiKey string) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (s *Service) Embed(ctx context.Context, text string) (Vector, error) {
	body, err := json.Marshal(request{
		Input: []string{text},
		Model: "voyage-4-lite",
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.voyageai.com/v1/embeddings",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling voyage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage returned status: %d", resp.StatusCode)
	}

	var result response
	// json.Unmarshal works on a []byte. It means you have to read the entire response body into memory first, then parse it.
	// json.NewDecoder works directly on a stream, in this case resp.Body. It reads and parses at the same time without loading everything into memory first.
	// Unmarshal is like reading an entire book before summarizing it. NewDecoder is like summarizing each chapter as you read it.
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return Vector(result.Data[0].Embedding), nil
}

func CosineSimilarity(a, b Vector) float32 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
