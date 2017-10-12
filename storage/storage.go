package storage

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var ErrNotFound = errors.New("not found")

type storageMetrics struct {
	createdChunk *prometheus.Counter
}

func newStorageMetrics(r prometheus.Registerer) *storageMetrics {
	m := &storageMetrics{}
	return m
}

type Options struct {
	Retention time.Duration
}

type storage struct {
	dbPath  string
	logger  *zap.Logger
	options *Options
	metrics *storageMetrics

	latest      *model.Snapshot
	latestChunk Chunk

	mtx sync.RWMutex
}

func Open(path string, logger *zap.Logger, r prometheus.Registerer, opts *Options) (Storage, error) {
	dbPath := strings.TrimSuffix(path, "/")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err = os.MkdirAll(dbPath, 0755)
		if err != nil {
			return nil, err
		}
	}
	// TODO: load latest chunk

	return &storage{
		dbPath:  dbPath,
		logger:  logger,
		options: opts,
		metrics: newStorageMetrics(r),
	}, nil
}

func (s *storage) Add(snapshot *model.Snapshot) error {
	if snapshot == nil {
		s.logger.Error("Snapshot is nil")
		return nil
	}

	chunkID := ChunkID(snapshot.Timestamp)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	logger := s.logger.With(
		zap.Time("ts", snapshot.Timestamp),
		zap.Int64("chunkID", chunkID))

	old := s.latest != nil && s.latest.Timestamp.After(snapshot.Timestamp)
	if !old {
		s.latest = snapshot
	}

	if s.latestChunk == nil {
		s.latestChunk = NewChunk(chunkID)
	}

	switch {
	case s.latestChunk.ID() == chunkID:
		if err := s.latestChunk.Add(snapshot); err != nil {
			logger.Error("Failed to add a new snapshot into a chunk", zap.Error(err))
		}

	case s.latestChunk.ID() < chunkID:
		s.saveChunk(s.latestChunk)
		s.latestChunk = NewChunk(chunkID)
		s.latestChunk.Add(snapshot)

	default:
		logger.Warn("Unabled to add too old snapshot")
	}

	return nil
}

func (s *storage) GetChunk(chunkID int64) (Chunk, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.latestChunk != nil {
		s.logger.Info("GetChunk", zap.Any("latestChunkID", s.latestChunk.ID), zap.Int64("chunkID", chunkID))
	} else {
		s.logger.Info("GetChunk (latestChunk is nil)", zap.Int64("chunkID", chunkID))
	}

	if s.latestChunk != nil && s.latestChunk.ID() == chunkID {
		return s.latestChunk.Clone(), nil
	}
	return s.loadChunk(chunkID)
}

func (s *storage) GetLatestSnapshot() (*model.Snapshot, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.latest == nil {
		return nil, ErrNotFound
	}
	return s.latest, nil
}

func (s *storage) Close() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.latestChunk != nil {
		s.saveChunk(s.latestChunk)
	}
	return nil
}

func (s *storage) saveChunk(chunk Chunk) {
	chunk.SetCompleted()
	data, err := chunk.Marshal()
	if err != nil {
		s.logger.Error("Failed to marshal chunk",
			zap.Error(err),
			zap.Any("chunk", chunk))
		return
	}

	path := fmt.Sprintf("%s/%d.json", s.dbPath, chunk.ID())
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		s.logger.Error("Failed to write snapshot to disk", zap.Error(err))
	}
}

func (s *storage) loadChunk(chunkID int64) (Chunk, error) {
	path := fmt.Sprintf("%s/%d.json", s.dbPath, chunkID)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	chunk := NewChunk(chunkID)
	if err = chunk.Unmarshal(data); err != nil {
		return nil, err
	}

	return chunk, nil
}
