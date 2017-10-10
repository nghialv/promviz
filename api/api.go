package api

import (
	"fmt"
	"net/http"

	"github.com/nghialv/promviz/cache"
	"github.com/nghialv/promviz/config"
	"github.com/nghialv/promviz/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"go.uber.org/zap"
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
}

type handler struct {
	logger *zap.Logger

	options  *Options
	config   *config.Config
	reloadCh chan chan error

	cache   cache.Cache
	querier storage.Querier
}

func NewHandler(logger *zap.Logger, r prometheus.Registerer, opts *Options) Handler {
	return &handler{
		logger:   logger,
		reloadCh: make(chan chan error),
		options:  opts,

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
	h.register(mux)
	handler := c.Handler(mux)
	return http.ListenAndServe(h.options.ListenAddress, handler)
}

func (h *handler) register(mux *http.ServeMux) {
	id := 0
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		id++
		h.logger.Info(fmt.Sprintf("%5d - New request", id))
		gd, err := h.querier.GetLatest()
		if err != nil {
			w.WriteHeader(401)
			return
		}
		h.logger.Info(fmt.Sprintf("%5d - Handled", id))
		w.Header().Set("Content-Type", "application/json")
		w.Write(gd.JSON)
	})
}

func (h *handler) Reload() <-chan chan error {
	return h.reloadCh
}

func (h *handler) Stop() error {
	return nil
}

func (h *handler) reload(w http.ResponseWriter, r *http.Request) {
	rc := make(chan error)
	h.reloadCh <- rc
	if err := <-rc; err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload config: %s", err), http.StatusInternalServerError)
	}
}
