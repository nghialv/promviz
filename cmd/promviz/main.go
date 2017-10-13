package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nghialv/promviz/api"
	"github.com/nghialv/promviz/cache"
	"github.com/nghialv/promviz/config"
	"github.com/nghialv/promviz/retrieval"
	"github.com/nghialv/promviz/storage"
	"github.com/nghialv/promviz/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "promviz",
		Name:      "config_last_reload_successful",
		Help:      "Whether the last configuration reload attempt was successful.",
	})
)

func main() {
	cfg := struct {
		configFile    string
		logLevel      string
		storagePath   string
		metricAddress string
		metricPath    string

		api       api.Options
		retrieval retrieval.Options
		cache     cache.Options
		storage   storage.Options
	}{}

	a := kingpin.New(filepath.Base(os.Args[0]), "The Promviz server")
	a.Version(version.Version)
	a.HelpFlag.Short('h')

	a.Flag("config.file", "Promviz configuration file path.").
		Default("promviz.yml").StringVar(&cfg.configFile)

	a.Flag("log.level", "The level of logging.").
		Default("info").StringVar(&cfg.logLevel)

	a.Flag("storage.path", "Base path for graph data storage.").
		Default("data/").StringVar(&cfg.storagePath)

	a.Flag("metric.listen-address", "Address to listen on for metrics.").
		Default(":9092").StringVar(&cfg.metricAddress)

	a.Flag("metric.path", "Path to output promviz inside metrics.").
		Default("/metrics").StringVar(&cfg.metricPath)

	a.Flag("api.listen-address", "Address to listen on for API requests.").
		Default(":9091").StringVar(&cfg.api.ListenAddress)

	a.Flag("retrieval.scrape-interval", "How frequently to scrape metrics from prometheuses.").
		Default("10s").DurationVar(&cfg.retrieval.ScrapeInterval)

	a.Flag("retrieval.scrape-timeout", "How long until a scrape request times out.").
		Default("8s").DurationVar(&cfg.retrieval.ScrapeTimeout)

	a.Flag("cache.size", "The maximum number of snapshots can be cached.").
		Default("100").IntVar(&cfg.cache.Size)

	a.Flag("storage.retention", "How long to retain graph data in the storage.").
		Default("24h").DurationVar(&cfg.storage.Retention)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Failed to parse arguments: %v\n", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	// TODO: log lever
	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(2)
	}
	defer logger.Sync()

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		prometheus.NewGoCollector(),
		configSuccess)

	storageReady := make(chan struct{})
	var db storage.Storage
	go func() {
		defer close(storageReady)
		var err error
		db, err = storage.Open(
			cfg.storagePath,
			logger.With(zap.String("component", "storage")),
			registry,
			&cfg.storage,
		)
		if err != nil {
			logger.Error("Failed to open db", zap.Error(err))
			os.Exit(1)
		}
	}()
	<-storageReady
	defer db.Close()
	cfg.api.Querier = db
	cfg.retrieval.Appender = db

	retriever := retrieval.NewRetriever(
		logger.With(zap.String("component", "retrieval")),
		registry,
		&cfg.retrieval,
	)

	if err := reloadConfig(cfg.configFile, logger, retriever); err != nil {
		logger.Error("Failed to run first reloading config", zap.Error(err))
	}

	go retriever.Run()
	defer retriever.Stop()

	cache := cache.NewCache(
		logger.With(zap.String("component", "cache")),
		registry,
		&cfg.cache,
	)
	defer cache.Reset()

	cfg.api.Cache = cache
	apiHandler := api.NewHandler(
		logger.With(zap.String("component", "api")),
		registry,
		&cfg.api,
	)

	logger.Info("Starting promviz", zap.String("info", version.String()))
	go apiHandler.Run()
	defer apiHandler.Stop()

	go func() {
		for {
			rc := <-apiHandler.Reload()
			if err := reloadConfig(cfg.configFile, logger, retriever); err != nil {
				logger.Error("Failed to reload config", zap.Error(err))
				rc <- err
			} else {
				rc <- nil
			}
		}
	}()

	errCh := make(chan error)
	go func() {
		mux := http.NewServeMux()
		mux.Handle(cfg.metricPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

		logger.Info("Starting metric server...",
			zap.String("address", cfg.metricAddress),
			zap.String("path", cfg.metricPath))

		if err := http.ListenAndServe(cfg.metricAddress, mux); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Warn("Received SIGTERM, exiting gracefully...")
	case err := <-errCh:
		logger.Error("Got an error from errCh, exiting gracefully", zap.Error(err))
	}
}

type Reloadable interface {
	ApplyConfig(*config.Config) error
}

func reloadConfig(path string, logger *zap.Logger, rl Reloadable) (err error) {
	logger.Info("Loading configuration file", zap.String("filepath", path))

	defer func() {
		if err != nil {
			configSuccess.Set(0)
		} else {
			configSuccess.Set(1)
		}
	}()

	cfg, err := config.LoadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to load configuration (--config.file=%s): %v", path, err)
	}
	err = rl.ApplyConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}
