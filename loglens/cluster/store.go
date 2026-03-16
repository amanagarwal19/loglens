package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"loglens/models"

	"github.com/redis/go-redis/v9"
)

const clusterKeyPrefix = "cluster:"

type Store struct {
	rdb *redis.Client
}

func NewStore(redisAddr string) *Store {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &Store{rdb: rdb}
}

// Save persists a cluster to Redis as JSON.
func (s *Store) Save(ctx context.Context, c *models.Cluster) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling cluster: %w", err)
	}
	return s.rdb.Set(ctx, clusterKeyPrefix+c.ID, data, 0).Err()
}

// LoadAll reads all clusters from Redis and returns them.
func (s *Store) LoadAll(ctx context.Context) ([]*models.Cluster, error) {
	keys, err := s.rdb.Keys(ctx, clusterKeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("fetching keys: %w", err)
	}

	var clusters []*models.Cluster
	for _, key := range keys {
		data, err := s.rdb.Get(ctx, key).Bytes()
		if err != nil {
			log.Printf("skipping key %s: %v", key, err)
			continue
		}

		var c models.Cluster
		if err := json.Unmarshal(data, &c); err != nil {
			log.Printf("skipping malformed cluster %s: %v", key, err)
			continue
		}
		clusters = append(clusters, &c)
	}
	return clusters, nil
}

// Reset deletes all clusters from Redis.
func (s *Store) Reset(ctx context.Context) error {
	keys, err := s.rdb.Keys(ctx, clusterKeyPrefix+"*").Result()
	if err != nil {
		return fmt.Errorf("fetching keys for reset: %w", err)
	}

	if len(keys) == 0 {
		log.Println("no clusters to reset")
		return nil
	}

	return s.rdb.Del(ctx, keys...).Err()
}
