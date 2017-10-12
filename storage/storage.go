package storage

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nghialv/promviz/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const chunkBlockLength = 2 * time.Minute // 5 * time.Hour

var (
	ErrNotFound = errors.New("Not found")
	ErrDBClosed = errors.New("DB already closed")
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
	dbDir   string
	logger  *zap.Logger
	options *Options
	metrics *storageMetrics

	latestSnapshot *model.Snapshot
	latestChunk    Chunk

	mtx    sync.RWMutex
	ctx    context.Context
	cancel func()
	doneCh chan struct{}
}

func Open(path string, logger *zap.Logger, r prometheus.Registerer, opts *Options) (Storage, error) {
	dbDir := strings.TrimSuffix(path, "/")
	if err := mkdirIfNotExist(dbDir); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	s := &storage{
		dbDir:   dbDir,
		logger:  logger,
		options: opts,
		metrics: newStorageMetrics(r),
		ctx:     ctx,
		cancel:  cancel,
		doneCh:  make(chan struct{}),
	}

	chunkID := ChunkID(time.Now())
	latestChunk, err := s.loadChunk(chunkID)
	if err != nil {
		s.logger.Warn("Unabled to load latest chunk from disk", zap.Error(err))
		latestChunk = NewChunk(chunkID)
	}

	latestChunk.SetCompleted(false)
	s.latestChunk = latestChunk
	go s.Run()

	return s, nil
}

func (s *storage) Add(snapshot *model.Snapshot) error {
	if snapshot == nil {
		s.logger.Error("Snapshot is nil")
		return nil
	}

	chunkID := ChunkID(snapshot.Timestamp)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	select {
	case <-s.ctx.Done():
		return ErrDBClosed
	default:
	}

	logger := s.logger.With(
		zap.Time("ts", snapshot.Timestamp),
		zap.Int64("chunkID", chunkID))

	old := s.latestSnapshot != nil && s.latestSnapshot.Timestamp.After(snapshot.Timestamp)
	if !old {
		s.latestSnapshot = snapshot
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
		s.latestChunk.SetCompleted(true)
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

	if s.latestChunk.ID() == chunkID {
		return s.latestChunk.Clone(), nil
	}
	return s.loadChunk(chunkID)
}

func (s *storage) GetLatestSnapshot() (*model.Snapshot, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.latestSnapshot == nil {
		return nil, ErrNotFound
	}
	return s.latestSnapshot, nil
}

func (s *storage) Run() {
	ticker := time.NewTicker(time.Minute) // TODO: move this to config
	defer func() {
		ticker.Stop()
		close(s.doneCh)
	}()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-ticker.C:
			s.retentionCutoff()
		}
	}
}

func (s *storage) Close() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	select {
	case <-s.doneCh:
		s.logger.Warn("Already closed")
		return ErrDBClosed
	default:
	}
	s.cancel()

	s.latestChunk.SetCompleted(true)
	err := s.saveChunk(s.latestChunk)
	if err != nil {
		s.logger.Error("Failed to save chunk to disk", zap.Error(err))
	}
	<-s.doneCh

	return err
}

func (s *storage) saveChunk(chunk Chunk) error {
	data, err := chunk.Marshal()
	if err != nil {
		s.logger.Error("Failed to marshal chunk",
			zap.Error(err),
			zap.Any("chunk", chunk))
		return err
	}

	bpath, cpath := chunkPath(s.dbDir, chunk.ID())
	if err := mkdirIfNotExist(bpath); err != nil {
		s.logger.Error("Failed to create block directory", zap.Error(err))
		return err
	}

	if err := ioutil.WriteFile(cpath, data, 0644); err != nil {
		s.logger.Error("Failed to write chunk to disk", zap.Error(err))
		return err
	}
	return nil
}

func (s *storage) loadChunk(chunkID int64) (Chunk, error) {
	_, cpath := chunkPath(s.dbDir, chunkID)
	data, err := ioutil.ReadFile(cpath)
	if err != nil {
		return nil, err
	}

	chunk := NewChunk(chunkID)
	if err = chunk.Unmarshal(data); err != nil {
		return nil, err
	}

	return chunk, nil
}

func chunkPath(dbDir string, chunkID int64) (blockPath string, chunkPath string) {
	bl := int64(chunkBlockLength.Seconds())
	blockTs := (chunkID / bl) * bl

	blockPath = fmt.Sprintf("%s/%d", dbDir, blockTs)
	chunkPath = fmt.Sprintf("%s/%d.json", blockPath, chunkID)
	return
}

func (s *storage) retentionCutoff() {
	mints := time.Now().Add(-s.options.Retention - chunkBlockLength).Unix()
	s.logger.Info("Will cutoff data", zap.Int64("mints", mints))

	if err := retentionCutoff(s.dbDir, mints); err != nil {
		s.logger.Error("Failed to cutoff old data", zap.Error(err))
	}
}

func retentionCutoff(dbDir string, mints int64) error {
	files, err := ioutil.ReadDir(dbDir)
	if err != nil {
		return err
	}
	var dirs []string

	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		ts, err := strconv.ParseInt(f.Name(), 10, 64)
		if err != nil {
			continue
		}
		if ts > mints {
			continue
		}
		dirs = append(dirs, filepath.Join(dbDir, f.Name()))
	}

	fmt.Println(dirs)
	fmt.Println(files)

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return nil
}

func mkdirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
