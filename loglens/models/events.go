package models

import (
	"loglens/embeddings"
	"time"
)

type LogEvent struct {
	ID        string            `json:"id"`
	Service   string            `json:"service"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}

type Cluster struct {
	ID        string            `json:"id"`
	Centroid  embeddings.Vector `json:"centroid"`
	Sample    LogEvent          `json:"sample"`
	Count     int               `json:"count"`
	FirstSeen time.Time         `json:"first_seen"`
	LastSeen  time.Time         `json:"last_seen"`
}
