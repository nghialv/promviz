package storage

import (
	"time"

	"github.com/nghialv/promviz/model"
)

// TODO: maybe this should be 5 minute
const ChunkLength = 30 * time.Second

type Chunk interface {
	ID() int64
	Add(*model.Snapshot) error
	FindBestSnapshot(time.Time) *model.Snapshot
	Clone() Chunk
	Completed()
	IsCompleted() bool
	Len() int
}

type chunk struct {
	id              int64             `json:"id"`
	sortedSnapshots []*model.Snapshot `json:"snapshots"`
	completed       bool              `json:"completed"`
}

func NewChunk(id int64) Chunk {
	return &chunk{
		id:              id,
		sortedSnapshots: make([]*model.Snapshot, 0),
	}
}

func ChunkID(ts time.Time) int64 {
	length := int64(ChunkLength / time.Second)
	return (ts.Unix() / length) * length
}

func (c *chunk) ID() int64 {
	return c.id
}

func (c *chunk) Add(snapshot *model.Snapshot) error {
	// TODO: append in order
	c.sortedSnapshots = append(c.sortedSnapshots, snapshot)
	return nil
}

func (c *chunk) FindBestSnapshot(ts time.Time) *model.Snapshot {
	if len(c.sortedSnapshots) == 0 {
		return nil
	}
	for i := len(c.sortedSnapshots) - 1; i >= 0; i-- {
		if ts.After(c.sortedSnapshots[i].Timestamp) {
			return c.sortedSnapshots[i]
		}
	}
	// TODO: should returns nil and load pre chunk to get more better one
	return c.sortedSnapshots[0]
}

func (c *chunk) Clone() Chunk {
	nc := NewChunk(c.id).(*chunk)
	for _, ss := range c.sortedSnapshots {
		snapshot := &model.Snapshot{}
		*snapshot = *ss
		nc.sortedSnapshots = append(nc.sortedSnapshots, snapshot)
	}
	return nc
}

func (c *chunk) Completed() {
	c.completed = true
}

func (c *chunk) IsCompleted() bool {
	return c.completed
}

func (c *chunk) Len() int {
	return len(c.sortedSnapshots)
}
