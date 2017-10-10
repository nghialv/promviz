package storage

import (
	"time"

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
	Get(string) (*model.Snapshot, error)
	GetLatest() (*model.Snapshot, error)
	GetKey(time.Time) string
}
