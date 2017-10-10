package model

import (
	"time"
)

type Chunk struct {
	ID              int64       `json:"id"`
	SortedSnapshots []*Snapshot `json:"snapshots"`
	Completed       bool        `json:"completed"`
}

func NewChunk(id int64) *Chunk {
	return &Chunk{
		ID:              id,
		SortedSnapshots: make([]*Snapshot, 0),
	}
}

func (c *Chunk) GetNearestSnapshot(ts time.Time) *Snapshot {
	if len(c.SortedSnapshots) == 0 {
		return nil
	}
	for i := len(c.SortedSnapshots) - 1; i >= 0; i-- {
		if ts.After(c.SortedSnapshots[i].Timestamp) {
			return c.SortedSnapshots[i]
		}
	}
	return c.SortedSnapshots[0]
}

func ChunkID(chunkLength time.Duration, ts time.Time) int64 {
	length := int64(chunkLength / time.Second)
	return (ts.Unix() / length) * length
}
