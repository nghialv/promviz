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

	a.Flag("metric.path", "Address to listen on for metrics.").
		Default("/metrics").StringVar(&cfg.metricPath)

	a.Flag("api.listen-address", "Address to listen on for API.").
		Default(":9091").StringVar(&cfg.api.ListenAddress)

	a.Flag("retrieval.scrape-interval", "").
		Default("10s").DurationVar(&cfg.retrieval.ScrapeInterval)

	a.Flag("retrieval.scrape-timeout", "").
		Default("10s").DurationVar(&cfg.retrieval.ScrapeTimeout)

	a.Flag("cache.size", "The max number of graph-data items can be cached.").
		Default("100").IntVar(&cfg.cache.Size)

	a.Flag("storage.retention", "How long to retain graph data in the storage.").
		Default("24h").DurationVar(&cfg.storage.Retention)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Failed to parse arguments: %v\n", err)
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(2)
	}
	defer logger.Sync()

	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())

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

	if err := reloadConfig(cfg.configFile, logger, retriever); err != nil {
		logger.Error("Failed to run first reloading config", zap.Error(err))
	}

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
		logger.Error("Failed to start web server, exiting gracefully", zap.Error(err))
	}
}

type Reloadable interface {
	ApplyConfig(*config.Config) error
}

func reloadConfig(path string, logger *zap.Logger, rl Reloadable) error {
	logger.Info("Loading configuration file", zap.String("filepath", path))

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
