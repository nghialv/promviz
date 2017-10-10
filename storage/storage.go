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
	latest *model.Snapshot
	mtx    sync.RWMutex
}

func Open(path string, logger *zap.Logger, r prometheus.Registerer, opts *Options) (Storage, error) {
	return &storage{
		logger: logger,
	}, nil
}

func (s *storage) Add(snapshot *model.Snapshot) error {
	if snapshot == nil {
		s.logger.Error("Snapshot is nil")
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.latest != nil && s.latest.Time.After(snapshot.Time) {
		s.logger.Info("Added a new snapshot but with wrong order", zap.Time("new", snapshot.Time), zap.Time("latestTime", snapshot.Time))
		return nil
	}

	s.logger.Info("Added a new snapshot", zap.Time("time", snapshot.Time))
	s.latest = snapshot
	return nil
}

func (s *storage) Get(t time.Time) (*model.Snapshot, error) {
	return nil, nil
}

func (s *storage) GetLatest() (*model.Snapshot, error) {
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
