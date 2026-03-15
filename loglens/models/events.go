package models

import "time"

type LogEvent struct {
	ID        string            `json:"id"`
	Service   string            `json:"service"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Meta      map[string]string `json:"meta,omitempty"`
}
