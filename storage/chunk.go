package storage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/nghialv/promviz/model"
)

const ChunkLength = 5 * time.Minute

var ErrAddToCompletedChunk = errors.New("Unabled to add a new snapshot into a completed chunk")

type Chunk interface {
	ID() int64
	SetCompleted(bool)
	IsCompleted() bool
	Len() int
	Clone() Chunk

	Add(*model.Snapshot) error

	Iterator() ChunkIterator

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type ChunkIterator interface {
	FindBestSnapshot(time.Time) *model.Snapshot
}

type chunk struct {
	TimestampID     int64             `json:"id"`
	SortedSnapshots []*model.Snapshot `json:"snapshots"`
	Completed       bool              `json:"completed"`
}

func NewChunk(id int64) Chunk {
	return &chunk{
		TimestampID:     id,
		SortedSnapshots: make([]*model.Snapshot, 0),
	}
}

func ChunkID(ts time.Time) int64 {
	length := int64(ChunkLength / time.Second)
	return (ts.Unix() / length) * length
}

func (c *chunk) ID() int64 {
	return c.TimestampID
}

func (c *chunk) SetCompleted(completed bool) {
	c.Completed = completed
}

func (c *chunk) IsCompleted() bool {
	return c.Completed
}

func (c *chunk) Len() int {
	return len(c.SortedSnapshots)
}

func (c *chunk) Clone() Chunk {
	nc := NewChunk(c.TimestampID).(*chunk)
	for _, ss := range c.SortedSnapshots {
		snapshot := &model.Snapshot{}
		*snapshot = *ss
		nc.SortedSnapshots = append(nc.SortedSnapshots, snapshot)
	}
	return nc
}

func (c *chunk) Add(snapshot *model.Snapshot) error {
	if c.Completed {
		return ErrAddToCompletedChunk
	}

	c.SortedSnapshots = append(c.SortedSnapshots, snapshot)
	for i := len(c.SortedSnapshots) - 1; i > 0; i-- {
		if c.SortedSnapshots[i].Timestamp.Before(c.SortedSnapshots[i-1].Timestamp) {
			tmp := c.SortedSnapshots[i-1]
			c.SortedSnapshots[i-1] = c.SortedSnapshots[i]
			c.SortedSnapshots[i] = tmp
		}
	}
	return nil
}

func (c *chunk) Iterator() ChunkIterator {
	return c
}

func (c *chunk) FindBestSnapshot(ts time.Time) *model.Snapshot {
	if len(c.SortedSnapshots) == 0 {
		return nil
	}
	for i := len(c.SortedSnapshots) - 1; i >= 0; i-- {
		if ts.After(c.SortedSnapshots[i].Timestamp) {
			return c.SortedSnapshots[i]
		}
	}
	// TODO: should returns nil and load pre chunk to get more better one
	return c.SortedSnapshots[0]
}

func (c *chunk) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c *chunk) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}
