package storage

import (
	"encoding/json"
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
	SetCompleted()
	IsCompleted() bool
	Len() int

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type Iterator interface {
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

func (c *chunk) Add(snapshot *model.Snapshot) error {
	// TODO: append in order
	c.SortedSnapshots = append(c.SortedSnapshots, snapshot)
	return nil
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

func (c *chunk) Clone() Chunk {
	nc := NewChunk(c.TimestampID).(*chunk)
	for _, ss := range c.SortedSnapshots {
		snapshot := &model.Snapshot{}
		*snapshot = *ss
		nc.SortedSnapshots = append(nc.SortedSnapshots, snapshot)
	}
	return nc
}

func (c *chunk) SetCompleted() {
	c.Completed = true
}

func (c *chunk) IsCompleted() bool {
	return c.Completed
}

func (c *chunk) Len() int {
	return len(c.SortedSnapshots)
}

func (c *chunk) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c *chunk) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}
