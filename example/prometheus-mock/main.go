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
	numberOfMeshConnections               = 100
	updateRate              time.Duration = 2     //number of seconds between updates
	maxRequestsPerSecond                  = 2000 // maximum number of requests per second for an application

	connections = []meshConnection{}

	clusters = []string{
		"us-east-1",
		"us-west-2",
		"ap-southeast-1",
		"eu-west-1",
	}

	httpCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "istio_request_count",
		Help: "Total number of requests",
	},
		[]string{
			"source_service",
			"source_version",
			"destination_service",
			"destination_version",
			"cluster",
			"response_code",
		},
	)

	httpServiceNames = []service{
		service{"istio-ingressgateway", []string{"v1"}},
		service{"parrot", []string{"v5"}},
		service{"penguin", []string{"v1", "v2"}},
		service{"toucan", []string{"v1"}},
		service{"cuckoos", []string{"v3"}},
		service{"woodpecker", []string{"v1"}},
		service{"passerine", []string{"v1", "v2"}},
		service{"albatross", []string{"v1"}},
		service{"anatidae", []string{"v6"}},
		service{"gregatidae", []string{"v4", "v5"}},
		service{"flowl", []string{"v1"}},
		service{"grouse", []string{"v1"}},
		service{"kiwis", []string{"v9"}},
		service{"swift", []string{"v1", "v2"}},
		service{"ightjar", []string{"v3"}},
		service{"guineafowl", []string{"v1", "v2"}},
		service{"turaco", []string{"v1"}},
		service{"neognathae", []string{"v3", "v4"}},
		service{"spoonbill", []string{"v11"}},
		service{"beeeater", []string{"v2"}},
		service{"shorebirds", []string{"v7", "v8"}},
	}
)

type meshConnection struct {
	sourceIndex        int
	destinationIndex   int
	splitPercent       int // 0-100 rps split percent from vx to vy
	requestsPerSeconds int // number of requests per second
	surge              int // max percent of rate that can change +x%
	errorRate          int //percent of errors
	total500           []float64
	total200           []float64
	total401           []float64
	total305           []float64
}
type service struct {
	name     string
	versions []string
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error)

	rg := prometheus.NewRegistry()
	rg.MustRegister(
		httpCounter,
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
		ticker := time.NewTicker(updateRate * time.Second)
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

	// generate mesh connections
	for i := 0; i < numberOfMeshConnections; i++ {
		sourceIndex, destinationIndex := generateIndices()
		errRate := rand.Intn(20) // max 20% error rate
		rps := rand.Intn(maxRequestsPerSecond)
		splitPercent := rand.Intn(100)
		surge := rand.Intn(20) // 20% surge in requests
		connection := meshConnection{
			sourceIndex:        sourceIndex,
			destinationIndex:   destinationIndex,
			errorRate:          errRate,
			requestsPerSeconds: rps,
			splitPercent:       splitPercent,
			surge:              surge,
			total200:           []float64{0.0, 0.0},
			total500:           []float64{0.0, 0.0},
			total401:           []float64{0.0, 0.0},
			total305:           []float64{0.0, 0.0},
		}
		connections = append(connections, connection)
	}
}

func generateIndices() (sourceIndex int, destinationIndex int) {
	sourceIndex = rand.Intn(len(httpServiceNames))
	destinationIndex = rand.Intn(len(httpServiceNames))
	if (sourceIndex == destinationIndex) || connectionAlreadyMade(sourceIndex, destinationIndex) {
		// keep looking for a new index if they are equal or already exist, recursive
		return generateIndices()
	}

	return sourceIndex, destinationIndex
}

func connectionAlreadyMade(sourceIndex int, destinationIndex int) bool {
	for _, c := range connections {
		if c.sourceIndex == sourceIndex && c.destinationIndex == destinationIndex {
			return true
		}
	}
	return false
}

func updateData() {
	fmt.Println("update metrics")

	for _, c := range connections {
		generateHttpMetric(&c)

	}
	//TODO generate istio-ingressgateway metrics everytime
}

func generateHttpMetric(connection *meshConnection) {

	for _, cluster := range clusters {

		//we need to generate requests for each cluster. we will randomize for each
		total := float64(connection.requestsPerSeconds * int(updateRate))
		surgePercent := float64(connection.surge) / float64(100)
		surge := float64(connection.requestsPerSeconds) * surgePercent

		total += surge

		rps500 := total * float64(connection.errorRate) / float64(100)
		rps200 := (total - rps500) * 0.9
		rps401 := (total - rps500) * 0.01
		rps305 := (total - rps500) * 0.09

		sourceService := httpServiceNames[connection.sourceIndex]
		destinationService := httpServiceNames[connection.destinationIndex]

		var splitRate []float64
		if len(destinationService.versions) == 1 {
			splitRate = []float64{float64(100)}
		} else if len(destinationService.versions) == 2 {

			splitRate = []float64{
				float64(1) - float64(connection.splitPercent/100),
				float64(connection.splitPercent),
			}
		} else {
			panic("Either 1 or 2 detination versions are required")
		}

		//there will only ever be 2 destination versions for now, we will split the requests based on that
		for i, s := range splitRate {

			connection.total200[i] += s * rps200
			connection.total305[i] += s * rps305
			connection.total401[i] += s * rps401
			connection.total500[i] += s * rps500

			httpCounter.WithLabelValues(sourceService.name, sourceService.versions[0], destinationService.name, destinationService.versions[i], cluster, "200").Set(connection.total200[i])
			httpCounter.WithLabelValues(sourceService.name, sourceService.versions[0], destinationService.name, destinationService.versions[i], cluster, "500").Set(connection.total500[i])
			httpCounter.WithLabelValues(sourceService.name, sourceService.versions[0], destinationService.name, destinationService.versions[i], cluster, "401").Set(connection.total401[i])
			httpCounter.WithLabelValues(sourceService.name, sourceService.versions[0], destinationService.name, destinationService.versions[i], cluster, "305").Set(connection.total305[i])
		}
	}
}

// func generateClusterMetric() {
// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-1", "200").Set(1000)
// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-1", "500").Set(10)

// 	clusterGauge.WithLabelValues("demo-cluster-1", "demo-cluster-2", "200").Set(100)
// 	clusterGauge.WithLabelValues("demo-cluster-1", "demo-cluster-2", "500").Set(1)

// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-2", "200").Set(500)
// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-2", "500").Set(5)

// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-3", "200").Set(100)
// 	clusterGauge.WithLabelValues("INTERNET", "demo-cluster-3", "500").Set(5)
// }
