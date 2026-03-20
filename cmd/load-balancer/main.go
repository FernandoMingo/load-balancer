package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"FernandoMingo/load-balancer/pkg/loadbalancer"
	"FernandoMingo/load-balancer/pkg/middleware"
)

type limiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients         = make(map[string]*limiterClient)
	mu              sync.Mutex
	cleanUpInterval = time.Minute * 5
)

// removes inactive clients
func cleanupClients() {
	for {
		time.Sleep(cleanUpInterval)
		mu.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > cleanUpInterval {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

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

	go cleanupClients()

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
		Handler: middleware.RecoveryMiddleware(middleware.
			LoggingMiddleware(middleware.RateLimitMiddleware(mux), log.Default())),
	}

	go lb.StartHealthCheck(ctx)

	log.Printf("Load balancer started at http://localhost:%d", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
