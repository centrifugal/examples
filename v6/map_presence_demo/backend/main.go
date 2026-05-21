// Massive map_clients presence demo for Centrifugo.
//
// Two scales served from a single Centrifugo node:
//   - 100k members, channel clients_100k:massive, viewer at /100k.html
//   - 1M  members, channel clients_1m:massive,    viewer at /1m.html
//
// Both populations are populated and churned synthetically via
// Centrifugo's HTTP API (`map_publish` / `map_remove`) — no real
// WebSocket clients are simulated. A single browser tab subscribes to a
// presence channel and watches the entire population synchronize over
// the protocol.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	channel100k = "clients_100k:massive"
	channel1m   = "clients_1m:massive"

	pool100k = 320 * 320   // 102,400
	pool1m   = 1024 * 1024 // 1,048,576

	count100k = 100000
	count1m   = 1000000

	churn100k = 200
	churn1m   = 1000
)

func main() {
	apiURL := envOr("CENTRIFUGO_API_URL", "http://localhost:8000/api")
	apiKey := envOr("CENTRIFUGO_API_KEY", "presence-demo-api-key")

	api := NewCentrifugoAPI(apiURL, apiKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("waiting for Centrifugo at %s ...", apiURL)
	if err := api.WaitReady(ctx); err != nil {
		log.Fatalf("Centrifugo not reachable: %v", err)
	}
	log.Printf("Centrifugo ready, starting farms")

	go runPresenceFarm(ctx, api, presenceFarmConfig{
		Channel: channel100k, PoolSize: pool100k, InitialCount: count100k, ChurnPerSec: churn100k,
	})
	go runPresenceFarm(ctx, api, presenceFarmConfig{
		Channel: channel1m, PoolSize: pool1m, InitialCount: count1m, ChurnPerSec: churn1m,
	})

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
