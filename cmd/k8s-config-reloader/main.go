package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"
	fsnotify "gopkg.in/fsnotify.v1"
)

var (
	Version = ""
)

type Config struct {
	promvizConfigDir string
	promvizReloadURL string
	logLevel         string
}

func main() {
	cfg := Config{}

	a := kingpin.New(filepath.Base(os.Args[0]), "The promviz config reloader for k8s")
	a.Version(Version)
	a.HelpFlag.Short('h')

	a.Flag("config.promviz-config-dir", "The directory contains Promviz configuration file.").
		StringVar(&cfg.promvizConfigDir)

	a.Flag("config.promviz-reload-url", "The url to send reloading request.").
		StringVar(&cfg.promvizReloadURL)

	a.Flag("config.log-level", "The level of logging.").
		Default("info").StringVar(&cfg.logLevel)

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

	rl := newReloader(&cfg, logger.With(zap.String("component", "reloader")))
	go func() {
		if err := rl.Run(); err != nil {
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Warn("Received SIGTERM, exiting gracefully...")
		rl.Stop()
	}
}

type reloader struct {
	cfg    *Config
	logger *zap.Logger

	ctx    context.Context
	cancel func()
	doneCh chan struct{}
}

func newReloader(cfg *Config, logger *zap.Logger) *reloader {
	ctx, cancel := context.WithCancel(context.Background())
	return &reloader{
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		doneCh: make(chan struct{}),
	}
}

func (r *reloader) Run() error {
	defer close(r.doneCh)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.logger.Error("Failed to create a new file watcher.", zap.Error(err))
		return err
	}
	defer watcher.Close()

	r.logger.Info("Starting reloader...")
	r.Reload()

	err = watcher.Add(r.cfg.promvizConfigDir)
	if err != nil {
		r.logger.Error("Failed to add config volume to be watched", zap.Error(err))
		return err
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				if filepath.Base(event.Name) == "..data" {
					r.logger.Info("ConfigMap modified")
					r.Reload()
				}
			}
		case err := <-watcher.Errors:
			r.logger.Error("Got an error from watcher", zap.Error(err))

		case <-r.ctx.Done():
			return nil
		}
	}
}

func (r *reloader) Stop() {
	r.cancel()
	<-r.doneCh
}

func (r *reloader) Reload() {
	cb := backoff.WithContext(backoff.NewExponentialBackOff(), r.ctx)
	err := backoff.RetryNotify(r.reload, cb, func(err error, next time.Duration) {
		r.logger.Warn("Failed to reload promviz configuration",
			zap.Error(err),
			zap.Duration("next retry", next),
		)
	})

	if err != nil {
		r.logger.Error("Failed to reload promviz configuration", zap.Error(err))
	} else {
		r.logger.Info("Reloaded successfully")
	}
}

func (r *reloader) reload() error {
	req, err := http.NewRequest("POST", r.cfg.promvizReloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Received response code %d, expected 200", resp.StatusCode)
	}
	return nil
}
