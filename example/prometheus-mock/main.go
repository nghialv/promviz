package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
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
)

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error)

	rg := prometheus.NewRegistry()
	rg.MustRegister(
		//prometheus.NewGoCollector(),
		httpGauge,
		grpcGauge,
		redisGauge,
		mongodbGauge,
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
			Type: "mongodb",
		})
	}

	for _, s := range redisServiceNames {
		redisServices = append(redisServices, &service{
			Name: s,
			Type: "redis",
		})
	}

	for i, s := range httpServiceNames {
		connectedServices := []*service{}
		for _, cs := range grpcServiceNames {
			if i == 0 || rand.Intn(len(grpcServiceNames)/3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "grpc",
				})
			}
		}

		for _, cs := range mongodbServiceNames {
			if rand.Intn(3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "mongodb",
				})
			}
		}

		for _, cs := range redisServiceNames {
			if rand.Intn(4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "redis",
				})
			}
		}

		httpServices = append(httpServices, &service{
			Name:              s,
			Type:              "http",
			ConnectedServices: connectedServices,
		})
	}

	for _, s := range grpcServiceNames {
		connectedServices := []*service{}

		for _, cs := range grpcServiceNames {
			num := len(grpcServiceNames) * len(grpcServiceNames)
			if s != cs && rand.Intn((num/5)*4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "grpc",
				})
			}
		}

		for _, cs := range mongodbServiceNames {
			if rand.Intn(3) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "mongodb",
				})
			}
		}

		for _, cs := range redisServiceNames {
			if rand.Intn(4) == 0 {
				connectedServices = append(connectedServices, &service{
					Name: cs,
					Type: "redis",
				})
			}
		}

		grpcServices = append(grpcServices, &service{
			Name:              s,
			Type:              "grpc",
			ConnectedServices: connectedServices,
		})
	}
}

func updateData() {
	fmt.Println("update data")

	for _, s := range httpServices {
		total := 20*rand.Float64() + 80
		generateHttpMetric(s.Name, total, 0.01)

		for _, cs := range s.ConnectedServices {
			switch cs.Type {
			case "grpc":
				total = 10*rand.Float64() + 40
				generateGrpcMetric(cs.Name, s.Name, total, 0.01)

			case "redis":
				total = 10*rand.Float64() + 40
				generateRedisMetric(cs.Name, s.Name, total, 0.01)

			case "mongodb":
				total = 10*rand.Float64() + 40
				generateMongodbMetric(cs.Name, s.Name, total, 0.01)
			}
		}
	}
}

type service struct {
	Name              string
	Type              string
	ConnectedServices []*service
}

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
	rps401 := (total - rps500) * 0.09
	rps305 := (total - rps500) * 0.01

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
	rpsSucceeded := total * errrate
	rpsFailed := total - rpsSucceeded

	redisGauge.WithLabelValues(service, dbname, "succeeded").Set(rpsSucceeded)
	redisGauge.WithLabelValues(service, dbname, "failed").Set(rpsFailed)
}

func generateMongodbMetric(service, dbname string, total float64, errrate float64) {
	rpsSucceeded := total * errrate
	rpsFailed := total - rpsSucceeded

	mongodbGauge.WithLabelValues(service, dbname, "succeeded").Set(rpsSucceeded)
	mongodbGauge.WithLabelValues(service, dbname, "failed").Set(rpsFailed)
}
