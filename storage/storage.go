package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	logger  *zap.Logger
	latest  *model.Snapshot
	mtx     sync.RWMutex
	options *Options
}

func Open(path string, logger *zap.Logger, r prometheus.Registerer, opts *Options) (Storage, error) {
	return &storage{
		logger:  logger,
		options: opts,
	}, nil
}

func (s *storage) Add(snapshot *model.Snapshot) error {
	if snapshot == nil {
		s.logger.Error("Snapshot is nil")
		return nil
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.latest != nil && s.latest.Timestamp.After(snapshot.Timestamp) {
		s.logger.Info("Added a new snapshot but with wrong order", zap.Time("new", snapshot.Timestamp), zap.Time("latestTime", snapshot.Timestamp))
		return nil
	}

	s.logger.Info("Added a new snapshot", zap.Time("time", snapshot.Timestamp))
	s.latest = snapshot

	path := fmt.Sprintf("%s/%s.json", "/Users/a13705/Downloads/db", s.GetKey(snapshot.Timestamp))
	err := ioutil.WriteFile(path, snapshot.GraphJSON, 0644)
	if err != nil {
		s.logger.Error("Failed to write snapshot to disk", zap.Error(err))
		return err
	}
	return nil
}

func (s *storage) Get(key string) (*model.Snapshot, error) {
	path := fmt.Sprintf("%s/%s.json", "/Users/a13705/Downloads/db", key)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &model.Snapshot{
		GraphJSON: data,
	}, nil
}

func (s *storage) GetLatest() (*model.Snapshot, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if s.latest == nil {
		return nil, ErrNotFound
	}
	return s.latest, nil
}

func (s *storage) GetKey(ts time.Time) string {
	kts := (ts.Unix() / 10) * 10
	return fmt.Sprintf("%d", kts)
}

func (s *storage) Close() error {
	return nil
}
