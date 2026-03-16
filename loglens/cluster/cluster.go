package cluster

import (
	"context"
	"fmt"
	"log"
	"loglens/embeddings"
	"loglens/models"
	"sync"
	"time"
)

// Engine groups incoming log events into clusters based on semantic similarity.
type Engine struct {
	mu        sync.Mutex
	clusters  []*models.Cluster
	embedSvc  *embeddings.Service
	threshold float32
	store     *Store
}

func New(embedSvc *embeddings.Service, threshold float32, store *Store) *Engine {
	e := &Engine{
		embedSvc:  embedSvc,
		threshold: threshold,
		store:     store,
	}

	clusters, err := e.store.LoadAll(context.Background())
	if err != nil {
		fmt.Printf("Error loading clusters from store: %v\n", err)
	} else {
		e.clusters = clusters
		fmt.Printf("Loaded %d clusters from store\n", len(clusters))
	}
	return e
}

// rollingAvg shifts the centroid toward the new vector incrementally.
// This avoids storing all vectors and recomputing the average from scratch.
func rollingAvg(centroid, newVec embeddings.Vector, count int) embeddings.Vector {
	result := make(embeddings.Vector, len(centroid))
	for i := range centroid {
		result[i] = (centroid[i]*float32(count-1) + newVec[i]) / float32(count)
	}
	return result
}

// Add takes a log event, embeds it, and either merges it into an
// existing cluster or creates a new one.
func (e *Engine) Add(ctx context.Context, event models.LogEvent) (*models.Cluster, bool, error) {
	vec, err := e.embedSvc.Embed(ctx, event.Message)
	if err != nil {
		return nil, false, fmt.Errorf("embedding log message: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Find the most similar existing cluster
	best, bestSim := e.findNearest(vec)

	if best != nil && bestSim >= e.threshold {
		// Merge into existing cluster
		best.Count++
		best.LastSeen = event.Timestamp
		best.Centroid = rollingAvg(best.Centroid, vec, best.Count)
		e.store.Save(ctx, best) // Update cluster info from inmemory to redis store
		return best, false, nil
	}

	// Create a new cluster
	cluster := &models.Cluster{
		ID:        fmt.Sprintf("cluster-%d", time.Now().UnixNano()),
		Centroid:  vec,
		Sample:    event,
		Count:     1,
		FirstSeen: event.Timestamp,
		LastSeen:  event.Timestamp,
	}
	e.clusters = append(e.clusters, cluster)
	e.store.Save(ctx, cluster) // Add new cluster from inmemory to redis store
	return cluster, true, nil
}

// All returns a snapshot of all current clusters.
func (e *Engine) All() []*models.Cluster {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.clusters
}

// findNearest finds the cluster whose centroid is most similar to vec.
func (e *Engine) findNearest(vec embeddings.Vector) (*models.Cluster, float32) {
	var best *models.Cluster
	var bestSim float32

	for _, c := range e.clusters {
		sim := embeddings.CosineSimilarity(c.Centroid, vec)
		if sim > bestSim {
			bestSim = sim
			best = c
		}
	}
	return best, bestSim
}

// Reset clears all clusters from memory and Redis.
func (e *Engine) Reset(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.store.Reset(ctx); err != nil {
		return fmt.Errorf("resetting store: %w", err)
	}

	e.clusters = nil
	log.Println("all clusters reset")
	return nil
}
