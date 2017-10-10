package storage

import (
	"time"

	"github.com/nghialv/promviz/model"
)

type Storage interface {
	// Querier(ctx context.Context) (Querier, error)
	// Appender() (Appender, error)
	Appender
	Querier
	Close() error
}

type Appender interface {
	Add(*model.GraphData) error
}

type Querier interface {
	Get(time.Time) (*model.GraphData, error)
	GetLatest() (*model.GraphData, error)
}
