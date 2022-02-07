package retrieval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nghialv/promviz/config"
	"github.com/prometheus/client_golang/api"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"go.uber.org/zap"
)

type querier interface {
	Query(context.Context, string, string, time.Time) (prommodel.Value, error)
	Stop() error
}

type promClient struct {
	addr     string
	client   api.Client
	queryAPI prometheus.API
}

type prompool struct {
	clients map[string]*promClient
	mtx     sync.Mutex
}

func newQuerier(logger *zap.Logger, cfg *config.Config) (*prompool, error) {
	addrs := make(map[string]struct{}, 0)
	for _, conn := range cfg.GlobalLevel.Connections {
		addrs[conn.PrometheusURL] = struct{}{}
	}
	for _, cluster := range cfg.ClusterLevel {
		for _, conn := range cluster.Connections {
			addrs[conn.PrometheusURL] = struct{}{}
		}
		for _, notice := range cluster.NodeNotices {
			addrs[notice.PrometheusURL] = struct{}{}
		}
	}
	delete(addrs, "")

	pq := &prompool{
		clients: make(map[string]*promClient, len(addrs)),
	}

	for addr := range addrs {
		c, err := api.NewClient(api.Config{Address: addr})
		if err != nil {
			return nil, err
		}
		a := prometheus.NewAPI(c)
		pq.clients[addr] = &promClient{
			addr:     addr,
			client:   c,
			queryAPI: a,
		}
	}
	return pq, nil
}

func (pp *prompool) Query(ctx context.Context, addr string, query string, ts time.Time) (prommodel.Value, error) {
	pp.mtx.Lock()
	client, _ := pp.clients[addr]
	pp.mtx.Unlock()

	if client == nil {
		return nil, fmt.Errorf("Could not send a query to unknown prometheus addr (addr=%s)", addr)
	}
	value, _, err := client.queryAPI.Query(ctx, query, ts)
	return value, err
}

func (pp *prompool) Stop() error {
	return nil
}
