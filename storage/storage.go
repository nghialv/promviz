package storage

import (
	"encoding/json"
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

var (
	ErrNotFound = errors.New("not found")
)

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
	latestChunk *model.Chunk

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

	chunkID := model.ChunkID(model.ChunkLength, snapshot.Timestamp)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	old := s.latest != nil && s.latest.Timestamp.After(snapshot.Timestamp)
	if !old {
		s.latest = snapshot
	}

	if s.latestChunk == nil {
		s.latestChunk = model.NewChunk(chunkID)
	}

	switch {
	case s.latestChunk.ID == chunkID:
		// TODO: append in order
		s.latestChunk.SortedSnapshots = append(s.latestChunk.SortedSnapshots, snapshot)

	case s.latestChunk.ID < chunkID:
		s.saveChunk(s.latestChunk)
		s.latestChunk = model.NewChunk(chunkID)
		s.latestChunk.SortedSnapshots = append(s.latestChunk.SortedSnapshots, snapshot)

	default:
		s.logger.Warn("Unabled to add too old snapshot",
			zap.Time("timestamp", snapshot.Timestamp),
			zap.Int64("chunkID", chunkID),
			zap.Int64("latestChunkID", s.latestChunk.ID))
	}

	return nil
}

func (s *storage) GetChunk(chunkID int64) (*model.Chunk, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.latestChunk != nil {
		s.logger.Info("GetChunk", zap.Any("latestChunkID", s.latestChunk.ID), zap.Int64("chunkID", chunkID))
	} else {
		s.logger.Info("GetChunk (latestChunk is nil)", zap.Int64("chunkID", chunkID))
	}

	if s.latestChunk != nil && s.latestChunk.ID == chunkID {
		c := model.NewChunk(chunkID)
		for _, ss := range s.latestChunk.SortedSnapshots {
			snapshot := &model.Snapshot{}
			*snapshot = *ss
			c.SortedSnapshots = append(c.SortedSnapshots, snapshot)
		}
		return c, nil
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

func (s *storage) saveChunk(chunk *model.Chunk) {
	path := fmt.Sprintf("%s/%d.json", s.dbPath, chunk.ID)
	chunk.Completed = true
	data, err := json.Marshal(chunk)
	if err != nil {
		s.logger.Error("Failed to marshal chunk",
			zap.Error(err),
			zap.Any("chunk", chunk))
		return
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		s.logger.Error("Failed to write snapshot to disk", zap.Error(err))
		return
	}
}

func (s *storage) loadChunk(chunkID int64) (*model.Chunk, error) {
	path := fmt.Sprintf("%s/%d.json", s.dbPath, chunkID)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	chunk := model.NewChunk(chunkID)

	err = json.Unmarshal(data, chunk)
	if err != nil {
		return nil, err
	}
	return chunk, nil
}
