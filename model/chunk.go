package model

import (
	"time"
)

type Chunk struct {
	Snapshots []*Snapshot `json:"snapshots"`
}

func (c *Chunk) GetNearestSnapshot(ts time.Time) *Snapshot {
	return nil
}
