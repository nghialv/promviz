package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const metricsFile = "/metrics.txt"

func main() {
	metrics, err := os.Open(metricsFile)
	if err != nil {
		panic(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("new metrics request")
		io.Copy(w, metrics)
	})

	go func() {
		addr := ":30001"
		fmt.Printf("Http server is running on %s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-sigCh:
		fmt.Println("Stopped after receiving SIGTERM signal")
	case err := <-errCh:
		fmt.Printf("Stopped due to error %s\n", err.Error())
	}
}
