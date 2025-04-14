package main

import (
	"log"

	"github.com/arleyar/go-record-signer/internal/db"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/messaging"
)

func main() {
	log.Println("Starting Record Worker")

	cfg := config.LoadConfig()

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	natsClient, err := messaging.New(cfg.NatsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	log.Printf("Connected to database and NATS")
	log.Println("Worker shutting down")
}
