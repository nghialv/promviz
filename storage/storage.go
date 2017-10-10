package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	ErrNotFound = errors.New("not found")
)

type storageMetrics struct {
}

type Options struct {
	Retention time.Duration
}

type storage struct {
	logger *zap.Logger
	latest *model.GraphData
	mtx    sync.RWMutex
}

func Open(path string, logger *zap.Logger, r prometheus.Registerer, opts *Options) (Storage, error) {
	return &storage{
		logger: logger,
	}, nil
}

func (s *storage) Add(gd *model.GraphData) error {
	if gd == nil {
		s.logger.Error("graphdata is nil")
		return nil
	}
	s.logger.Info("Added a new graph data", zap.Time("time", gd.Time))

	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.latest = gd
	return nil
}

func (s *storage) Get(t time.Time) (*model.GraphData, error) {
	return nil, nil
}

func (s *storage) GetLatest() (*model.GraphData, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if s.latest == nil {
		return nil, ErrNotFound
	}
	return s.latest, nil
}

func (s *storage) Close() error {
	return nil
}
