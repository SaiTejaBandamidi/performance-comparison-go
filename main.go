package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("starting performance-comparison-go application")

	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("failed to load config.json: %v", err)
	}

	db, err := NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	log.Println("connected to postgres successfully")

	metricsStore := NewMetricsStore(db)
	benchmarkService := NewBenchmarkService(metricsStore)

	restServer := StartRESTServer(benchmarkService)
	go func() {
		log.Println("REST server listening on :8000")
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server error: %v", err)
		}
	}()

	graphQLServer := StartGraphQLServer(benchmarkService)
	go func() {
		log.Println("GraphQL server listening on :8080")
		if err := graphQLServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("GraphQL server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := restServer.Shutdown(ctx); err != nil {
		log.Printf("error shutting down REST server: %v", err)
	}

	if err := graphQLServer.Shutdown(ctx); err != nil {
		log.Printf("error shutting down GraphQL server: %v", err)
	}

	log.Println("application stopped")
}
