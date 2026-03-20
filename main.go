package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	var serverList string
	var port int
	flag.StringVar(&serverList, "backends", "",
		"List the backedns to load balance separated by commas")
	flag.IntVar(&port, "port", 3030,
		"Specify the port in which requests are received")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	// parse the servers
	tokens := strings.Split(serverList, ",")
	for _, tok := range tokens {
		serverURL, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverURL)
		// TODO: initial server aliveness check
		serverPool.AddBackend(&Backend{
			URL:          serverURL,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured Server: %s\n", serverURL)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	log.Printf("Load balancer started at: %d\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
