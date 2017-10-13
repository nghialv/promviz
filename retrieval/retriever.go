package retrieval

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/nghialv/promviz/config"
	"github.com/nghialv/promviz/storage"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	namespace = "promviz"
	subsystem = "retriver"

	ErrConfigNotSet = errors.New("Config has been not set")
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

type retrieverMetrics struct {
	ops       *prometheus.CounterVec
	opLatency *prometheus.SummaryVec
}

func newRetrieverMetrics(r prometheus.Registerer) *retrieverMetrics {
	m := &retrieverMetrics{
		ops: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "ops_total",
			Help:      "Total number of handled ops.",
		},
			[]string{"op", "status"},
		),
		opLatency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "op_latency_seconds",
			Help:      "Latency for handling op.",
		},
			[]string{"op", "status"},
		),
	}

	if r != nil {
		r.MustRegister(
			m.ops,
			m.opLatency,
		)
	}
	return m
}

type retriever struct {
	logger  *zap.Logger
	options *Options
	config  *config.Config
	metrics *retrieverMetrics

	appender storage.Appender
	querier  querier

	mtx    sync.RWMutex
	ctx    context.Context
	cancel func()
	doneCh chan struct{}
}

func NewRetriever(logger *zap.Logger, r prometheus.Registerer, opts *Options) Retriever {
	ctx, cancel := context.WithCancel(context.Background())

	return &retriever{
		logger:  logger,
		options: opts,
		metrics: newRetrieverMetrics(r),

		appender: opts.Appender,

		ctx:    ctx,
		cancel: cancel,
		doneCh: make(chan struct{}),
	}
}

func (r *retriever) Run() {
	r.logger.Info("Starting retriever...")
	defer close(r.doneCh)

	retrieve := func() {
		r.logger.Info("Retrieve prometheus data and generate graph data")
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(r.options.ScrapeTimeout))
		r.retrieve(ctx)
		cancel()
	}
	retrieve()

	ticker := time.NewTicker(r.options.ScrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return

		case <-ticker.C:
			retrieve()
		}
	}
}

func (r *retriever) Stop() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	select {
	case <-r.doneCh:
		r.logger.Warn("Already stopped")
		return
	default:
	}

	r.logger.Info("Stopping retriever...")
	r.cancel()
	<-r.doneCh
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

func (r *retriever) retrieve(ctx context.Context) (err error) {
	defer track(r.metrics, "Retrieve")(&err)

	r.mtx.RLock()
	cfg := r.config
	querier := r.querier
	r.mtx.RUnlock()

	if cfg == nil {
		r.logger.Warn("Config has not been set")
		return ErrConfigNotSet
	}

	g := &generator{
		logger:  r.logger,
		cfg:     cfg,
		querier: querier,
	}

	snapshot, err := g.generateSnapshot(ctx, time.Now())
	if err != nil {
		r.logger.Error("Failed to generate graph data", zap.Error(err))
		return
	}

	err = r.appender.Add(snapshot)
	if err != nil {
		r.logger.Error("Failed to add snapshot to storage", zap.Error(err))
		return
	}

	return
}

func track(metrics *retrieverMetrics, op string) func(*error) {
	start := time.Now()
	return func(err *error) {
		s := strconv.FormatBool(*err == nil)
		metrics.ops.WithLabelValues(op, s).Inc()
		metrics.opLatency.WithLabelValues(op, s).Observe(time.Since(start).Seconds())
	}
}
