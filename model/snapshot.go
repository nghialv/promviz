package model

import (
	"time"
)

type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	GraphJSON string    `json:"graphJSON"`
}
