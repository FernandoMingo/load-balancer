package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"FernandoMingo/load-balancer/pkg/loadbalancer"
	"FernandoMingo/load-balancer/pkg/middleware"
)

func main() {
	var serverList string
	var port int

	flag.StringVar(
		&serverList,
		"backends",
		"",
		"List of backend URLs separated by commas",
	)
	flag.IntVar(
		&port,
		"port",
		3030,
		"Port in which requests are received",
	)
	flag.Parse()

	lb, err := loadbalancer.New(strings.Split(serverList, ","), 20*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.Handle("/", lb)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: middleware.LoggingMiddleware(middleware.
			RateLimitMiddleware(middleware.RecoveryMiddleware(mux)), log.Default()),
	}

	go lb.StartHealthCheck(ctx)

	log.Printf("Load balancer started at: %d\n", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
