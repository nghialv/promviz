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
	"github.com/rs/cors"
	"go.uber.org/zap"
)

var (
	namespace = "promviz"
	subsystem = "api"
)

type Handler interface {
	Run() error
	Stop() error
	Reload() <-chan chan error
}

type Options struct {
	ListenAddress string
	Cache         cache.Cache
	Querier       storage.Querier
}

type apiMetrics struct {
	requests *prometheus.CounterVec
	latency  *prometheus.SummaryVec
}

func newApiMetrics(r prometheus.Registerer) *apiMetrics {
	m := &apiMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of handled requests.",
		},
			[]string{"path", "status"},
		),
		latency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "latency_seconds",
			Help:      "Latency for handling request.",
		},
			[]string{"path"},
		),
	}
	if r != nil {
		r.MustRegister(
			m.requests,
			m.latency,
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

func (h *handler) Run() error {
	h.logger.Info("Start listening for connections", zap.String("address", h.options.ListenAddress))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.getGraphHandler)
	mux.HandleFunc("/reload", h.reloadHandler)

	return http.ListenAndServe(h.options.ListenAddress, c.Handler(mux))
}

func (h *handler) Reload() <-chan chan error {
	return h.reloadCh
}

func (h *handler) Stop() error {
	h.logger.Info("Stopping api server...")
	return nil
}

func (h *handler) reloadHandler(w http.ResponseWriter, r *http.Request) {
	rc := make(chan error)
	h.reloadCh <- rc
	if err := <-rc; err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload config: %s", err), http.StatusInternalServerError)
	}
}

func (h *handler) getGraphHandler(w http.ResponseWriter, req *http.Request) {
	getSnapshot := h.querier.GetLatest
	query := req.URL.Query()
	offsets := query["offset"]

	if len(offsets) != 0 {
		offset, err := strconv.Atoi(offsets[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid offset (%s): %s", offsets[0], err), http.StatusBadRequest)
			return
		}

		ts := time.Now().Add(time.Duration(-offset) * time.Second)
		key := h.querier.GetKey(ts)

		getSnapshot = func() (*model.Snapshot, error) {
			if c := h.cache.Get(key); c != nil {
				return c, nil
			}

			h.logger.Warn("Cache miss",
				zap.Time("ts", ts),
				zap.String("key", key))

			snapshot, err := h.querier.Get(key)
			if err == nil {
				h.cache.Put(key, snapshot)
			}
			return snapshot, err
		}
	}

	snapshot, err := getSnapshot()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get snapshot: %s", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(snapshot.GraphJSON)
}
