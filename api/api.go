package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nghialv/promviz/cache"
	"github.com/nghialv/promviz/model"
	"github.com/nghialv/promviz/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

var (
	namespace = "promviz"
	subsystem = "api"
)

type Handler interface {
	Run(prometheus.Gatherer) error
	Stop() error
	Reload() <-chan chan error
}

type Options struct {
	ListenPort int
	Cache      cache.Cache
	Querier    storage.Querier
}

type apiMetrics struct {
	requests         *prometheus.CounterVec
	latency          *prometheus.SummaryVec
	snapshotNotFound prometheus.Counter
}

func newApiMetrics(r prometheus.Registerer) *apiMetrics {
	m := &apiMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of handled requests.",
		},
			[]string{"handler", "status"},
		),
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "latency_seconds",
			Help:      "Latency for handling request.",
		},
			[]string{"handler", "status"},
		),
		snapshotNotFound: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "snapshot_not_found",
			Help:      "Total number of times unabled to find snapshot.",
		}),
	}
	if r != nil {
		r.MustRegister(
			m.requests,
			m.latency,
			m.snapshotNotFound,
		)
	}
	return m
}

type handler struct {
	logger  *zap.Logger
	metrics *apiMetrics

	options  *Options
	reloadCh chan chan error

	cache   cache.Cache
	querier storage.Querier
}

func NewHandler(logger *zap.Logger, r prometheus.Registerer, opts *Options) Handler {
	return &handler{
		logger:  logger,
		metrics: newApiMetrics(r),

		options:  opts,
		reloadCh: make(chan chan error),

		cache:   opts.Cache,
		querier: opts.Querier,
	}
}

func (h *handler) Run(g prometheus.Gatherer) error {
	addr := fmt.Sprintf(":%d", h.options.ListenPort)
	h.logger.Info("Start listening for incoming connections", zap.String("address", addr))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/graph", h.getGraphHandler)
	mux.HandleFunc("/reload", h.reloadHandler)
	mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))

	return http.ListenAndServe(addr, c.Handler(mux))
}

func (h *handler) Reload() <-chan chan error {
	return h.reloadCh
}

func (h *handler) Stop() error {
	h.logger.Info("Stopping api server...")
	return nil
}

func (h *handler) reloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("Invalid request method"), http.StatusNotImplemented)
		return
	}

	status := http.StatusOK
	defer track(h.metrics, "Reload")(&status)

	rc := make(chan error)
	h.reloadCh <- rc
	if err := <-rc; err != nil {
		status = http.StatusInternalServerError
		http.Error(w, fmt.Sprintf("Failed to reload config: %s", err), status)
		return
	}
	w.WriteHeader(status)
}

func (h *handler) getGraphHandler(w http.ResponseWriter, req *http.Request) {
	status := http.StatusOK
	defer track(h.metrics, "GetGraph")(&status)

	getSnapshot := h.querier.GetLatestSnapshot
	query := req.URL.Query()
	offsets := query["offset"]

	if len(offsets) != 0 {
		offset, err := strconv.Atoi(offsets[0])
		if err != nil {
			status = http.StatusBadRequest
			http.Error(w, fmt.Sprintf("Invalid offset (%s): %s", offsets[0], err), status)
			return
		}

		if offset > 0 {
			ts := time.Now().Add(time.Duration(-offset) * time.Second)
			chunkID := storage.ChunkID(ts)

			getSnapshot = func() (*model.Snapshot, error) {
				chunk := h.cache.Get(chunkID)
				if chunk == nil {
					h.logger.Warn("Cache miss",
						zap.Time("ts", ts),
						zap.Int64("chunkID", chunkID))

					chunk, err = h.querier.GetChunk(chunkID)
					if err != nil {
						return nil, err
					}
					if chunk.IsCompleted() {
						h.cache.Put(chunkID, chunk)
					}
				}

				snapshot := chunk.Iterator().FindBestSnapshot(ts)
				if snapshot != nil {
					return snapshot, nil
				}

				h.metrics.snapshotNotFound.Inc()
				h.logger.Warn("Unabled to find snapshot in chunk",
					zap.Time("ts", ts),
					zap.Int("chunkLength", chunk.Len()),
					zap.Any("chunk", chunk))

				return nil, fmt.Errorf("Not found snapshot")
			}
		}
	}

	snapshot, err := getSnapshot()
	if err != nil {
		status = http.StatusNotFound
		http.Error(w, fmt.Sprintf("Failed to get snapshot: %s", err), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(snapshot.GraphJSON))
}

func track(metrics *apiMetrics, handler string) func(*int) {
	start := time.Now()
	return func(status *int) {
		s := fmt.Sprintf("%d", *status)
		metrics.requests.WithLabelValues(handler, s).Inc()
		metrics.latency.WithLabelValues(handler, s).Observe(time.Since(start).Seconds())
	}
}
