package model

import (
	"time"
)

type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	GraphJSON []byte    `json:"graphJSON"`
}
