package main

import (
	"context"
	"log"
	"time"

	"github.com/arleyar/go-record-signer/internal/db"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/messaging"
)

func main() {
	log.Println("Starting Record Dispatcher")

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

	log.Printf("Connected to database and NATS, batch size: %d", cfg.BatchSize)

	for {
		hasRecords := dispatchBatch(database, natsClient, cfg.BatchSize)
		if !hasRecords {
			log.Println("No more pending records, exiting")
			break
		}
	}

	log.Println("Dispatcher completed successfully")
}

func dispatchBatch(database *db.DB, natsClient *messaging.NATSClient, batchSize int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	records, err := database.GetPendingRecords(ctx, batchSize)
	if err != nil {
		log.Printf("Error getting pending records: %v", err)
		return false
	}

	if len(records) == 0 {
		return false
	}

	if err = natsClient.PublishBatch(records); err != nil {
		log.Printf("Error publishing batch: %v", err)
		return true
	}

	if err = database.UpdateRecordsToQueued(ctx, records); err != nil {
		log.Printf("Error updating records to queued: %v", err)
		return true
	}

	log.Printf("Published batch of %d records", len(records))
	return true
}
