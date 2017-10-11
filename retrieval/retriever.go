package retrieval

import (
	"context"
	"sync"
	"time"

	"github.com/nghialv/promviz/config"
	"github.com/nghialv/promviz/storage"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Retriever interface {
	Run()
	Stop()
	ApplyConfig(*config.Config) error
}

type Options struct {
	ScrapeInterval time.Duration
	ScrapeTimeout  time.Duration
	Appender       storage.Appender
}

type retriever struct {
	logger  *zap.Logger
	options *Options
	config  *config.Config

	appender storage.Appender
	querier  querier

	mtx    sync.RWMutex
	ctx    context.Context
	cancel func()
	done   chan struct{}
}

func NewRetriever(logger *zap.Logger, r prometheus.Registerer, opts *Options) Retriever {
	return &retriever{
		logger:   logger,
		options:  opts,
		appender: opts.Appender,
		done:     make(chan struct{}),
	}
}

func (r *retriever) Run() {
	r.logger.Info("Starting retriever...")
	r.mtx.Lock()
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.mtx.Unlock()

	defer close(r.done)
	ticker := time.NewTicker(r.options.ScrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return

		case <-ticker.C:
			r.mtx.RLock()
			cfg := r.config
			querier := r.querier
			r.mtx.RUnlock()

			r.logger.Info("Retrieves new data")
			if cfg == nil {
				r.logger.Warn("Config was not set")
				continue
			}

			g := &generator{
				logger:  r.logger,
				cfg:     cfg,
				querier: querier,
			}
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(r.options.ScrapeTimeout))
			snapshot, err := g.generateSnapshot(ctx, time.Now())
			cancel()
			if err != nil {
				r.logger.Error("Failed to generate graph data", zap.Error(err))
			} else {
				r.appender.Add(snapshot)
			}
		}
	}
}

func (r *retriever) Stop() {
	r.logger.Info("Stopping retriever...")
	r.mtx.Lock()
	r.cancel()
	r.mtx.Unlock()
	<-r.done
}

func (r *retriever) ApplyConfig(cfg *config.Config) error {
	q, err := newQuerier(r.logger, cfg)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.config = cfg
	r.querier = q
	r.logger.Info("Applied new configuration")

	return nil
}
