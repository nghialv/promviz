package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "status:http_requests_total:rate2m",
		Help: "Requests per second",
	},
		[]string{"service", "status"},
	)

	grpcGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "client_code:grpc_server_requests_total:rate2m",
		Help: "Requests per second",
	},
		[]string{"service", "client", "code"},
	)

	redisGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "status:redis_client_cmds_total:rate2m",
		Help: "Requests per second",
	},
		[]string{"service", "dbname", "status"},
	)

	mongodbGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "status:mongodb_client_ops_total:rate2m",
		Help: "Requests per second",
	},
		[]string{"service", "dbname", "status"},
	)

	clusterGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster:http_requests_total:rate2m",
		Help: "Requests per second",
	},
		[]string{"source", "target", "status"},
	)
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error)

	rg := prometheus.NewRegistry()
	rg.MustRegister(
		httpGauge,
		grpcGauge,
		redisGauge,
		mongodbGauge,
		clusterGauge,
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(rg, promhttp.HandlerOpts{}))

	go func() {
		addr := ":30001"
		fmt.Printf("Http server is running on %s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			errCh <- err
		}
	}()

	doneCh := make(chan struct{})
	defer close(doneCh)

	initilizeServices()
	generateClusterMetric()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-doneCh:
				return

			case <-ticker.C:
				updateData()
			}
		}
	}()

	select {
	case <-sigCh:
		fmt.Println("Stopped after receiving SIGTERM signal")
	case err := <-errCh:
		fmt.Printf("Stopped due to error %s\n", err.Error())
	}
}

func initilizeServices() {
	for _, s := range mongodbServiceNames {
		mongodbServices = append(mongodbServices, &service{
			Name: s,
			Type: ST_MONGODB,
		})
	}

	for _, s := range redisServiceNames {
		redisServices = append(redisServices, &service{
			Name: s,
			Type: ST_REDIS,
		})
	}

	for i, s := range httpServiceNames {
		connectedServices := []*service{}

		for _, cs := range grpcServiceNames {
			if i == 0 || rand.Intn(len(grpcServiceNames)/3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_GRPC,
				})
			}
		}

		for _, cs := range mongodbServiceNames {
			if rand.Intn(3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_MONGODB,
				})
			}
		}

		for _, cs := range redisServiceNames {
			if rand.Intn(4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_REDIS,
				})
			}
		}

		httpServices = append(httpServices, &service{
			Name:              s,
			Type:              ST_HTTP,
			ConnectedServices: connectedServices,
		})
	}

	for _, s := range grpcServiceNames {
		connectedServices := []*service{}

		for _, cs := range grpcServiceNames {
			if strings.Compare(s, cs) > 0 && rand.Intn(len(grpcServiceNames)*5/4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_GRPC,
				})
			}
		}

		for _, cs := range mongodbServiceNames {
			if rand.Intn(3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_MONGODB,
				})
			}
		}

		for _, cs := range redisServiceNames {
			if rand.Intn(4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: ST_REDIS,
				})
			}
		}

		grpcServices = append(grpcServices, &service{
			Name:              s,
			Type:              ST_GRPC,
			ConnectedServices: connectedServices,
		})
	}
}

func updateData() {
	fmt.Println("update metrics")

	index := 0
	for _, s := range httpServices {
		num := 100.0
		errRate := 0.0
		if index == 0 {
			errRate = 0.0125
			num = 500.0
		}
		if index == 4 {
			errRate = 0.00002
			num = 300.0
		}
		index++

		total := num*rand.Float64() + num
		generateHttpMetric(s.FullName(), total, errRate)

		for _, cs := range s.ConnectedServices {
			switch cs.Type {
			case ST_GRPC:
				total = 10*rand.Float64() + 40
				generateGrpcMetric(cs.FullName(), s.FullName(), total, 0.01)

			case ST_REDIS:
				total = 10*rand.Float64() + 40
				generateRedisMetric(s.FullName(), cs.FullName(), total, 0.0)

			case ST_MONGODB:
				total = 10*rand.Float64() + 40
				generateMongodbMetric(s.FullName(), cs.FullName(), total, 0.01)
			}
		}
	}

	index = 0
	for _, s := range grpcServices {
		for _, cs := range s.ConnectedServices {
			switch cs.Type {
			case ST_GRPC:
				total := 10*rand.Float64() + 40
				generateGrpcMetric(cs.FullName(), s.FullName(), total, 0.01)

			case ST_REDIS:
				errRate := 0.0
				if index == 0 {
					errRate = 0.0125
				}
				index++

				total := 10*rand.Float64() + 40
				generateRedisMetric(s.FullName(), cs.FullName(), total, errRate)

			case ST_MONGODB:
				total := 10*rand.Float64() + 40
				generateMongodbMetric(s.FullName(), cs.FullName(), total, 0.01)
			}
		}
	}
}

type service struct {
	Name              string
	Type              string
	ConnectedServices []*service
}

func (s *service) FullName() string {
	return fmt.Sprintf("%s-%s", s.Type, s.Name)
}

const (
	ST_HTTP    = "http"
	ST_GRPC    = "grpc"
	ST_REDIS   = "redis"
	ST_MONGODB = "mongodb"
)

var (
	httpServices    = []*service{}
	grpcServices    = []*service{}
	redisServices   = []*service{}
	mongodbServices = []*service{}

	httpServiceNames = []string{
		"sparrow",
		"hummingbird",
		"coraciiformes",
		"ciconfiiformes",
		"kingfisher",
	}
	grpcServiceNames = []string{
		"parrot",
		"penguin",
		"toucan",
		"cuckoos",
		"woodpecker",
		"passerine",
		"albatross",
		"anatidae",
		"anatidae",
		"gregatidae",
		"flowl",
		"grouse",
		"kiwis",
		"swift",
		"ightjar",
		"guineafowl",
		"turaco",
		"guineafowl",
		"neognathae",
		"spoonbill",
		"beeeater",
		"shorebirds",
	}
	redisServiceNames = []string{
		"swallow",
		"bulbul",
		"strork",
		"hornbill",
	}
	mongodbServiceNames = []string{
		"sandpiper",
		"goose",
		"plover",
	}
)

func generateHttpMetric(service string, total float64, errrate float64) {
	rps500 := total * errrate
	rps200 := (total - rps500) * 0.9
	rps401 := (total - rps500) * 0.01
	rps305 := (total - rps500) * 0.09

	httpGauge.WithLabelValues(service, "200").Set(rps200)
	httpGauge.WithLabelValues(service, "500").Set(rps500)
	httpGauge.WithLabelValues(service, "401").Set(rps401)
	httpGauge.WithLabelValues(service, "305").Set(rps305)
}

func generateGrpcMetric(service, client string, total float64, errrate float64) {
	rpsInternal := total * errrate
	rpsOK := (total - rpsInternal) * 0.9
	rpsNotFound := (total - rpsInternal) * 0.09
	rpsUnauthenticated := (total - rpsInternal) * 0.01

	grpcGauge.WithLabelValues(service, client, "OK").Set(rpsOK)
	grpcGauge.WithLabelValues(service, client, "Internal").Set(rpsInternal)
	grpcGauge.WithLabelValues(service, client, "NotFound").Set(rpsNotFound)
	grpcGauge.WithLabelValues(service, client, "Unauthenticated").Set(rpsUnauthenticated)
}

func generateRedisMetric(service, dbname string, total float64, errrate float64) {
	rpsFailed := total * errrate
	rpsSucceeded := total - rpsFailed

	redisGauge.WithLabelValues(service, dbname, "succeeded").Set(rpsSucceeded)
	redisGauge.WithLabelValues(service, dbname, "failed").Set(rpsFailed)
}

func generateMongodbMetric(service, dbname string, total float64, errrate float64) {
	rpsFailed := total * errrate
	rpsSucceeded := total - rpsFailed

	mongodbGauge.WithLabelValues(service, dbname, "succeeded").Set(rpsSucceeded)
	mongodbGauge.WithLabelValues(service, dbname, "failed").Set(rpsFailed)
}

func generateClusterMetric() {
	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-1", "200").Set(1000)
	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-1", "500").Set(10)

	clusterGauge.WithLabelValues("demo-cluster-1", "demo-cluster-2", "200").Set(100)
	clusterGauge.WithLabelValues("demo-cluster-1", "demo-cluster-2", "500").Set(1)

	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-2", "200").Set(500)
	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-2", "500").Set(5)

	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-3", "200").Set(100)
	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-3", "500").Set(5)
}
