package storage

import (
	"github.com/nghialv/promviz/model"
)

type Storage interface {
	Appender
	Querier
	Close() error
}

type Appender interface {
	Add(*model.Snapshot) error
}

type Querier interface {
	GetChunk(int64) (*model.Chunk, error)
	GetLatestSnapshot() (*model.Snapshot, error)
}
