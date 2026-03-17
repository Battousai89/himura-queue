package models

import "time"

type Message struct {
	ID        uint64        `json:"id"`
	Queue     string        `json:"queue"`
	Payload   []byte        `json:"payload"`
	Priority  int           `json:"priority"`
	Delay     time.Duration `json:"delay,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at,omitempty"`
}
