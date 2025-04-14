package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arleyar/go-record-signer/internal/db"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/crypto"
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

	key, err := cfg.GetEncryptionKey()
	if err != nil {
		log.Fatalf("Failed to get encryption key: %v", err)
	}

	encryptor, err := crypto.NewKeyEncryptor(key)
	if err != nil {
		log.Fatalf("Failed to create key encryptor: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sub, err := natsClient.SubscribeBatch(func(ctx context.Context, batch *messaging.BatchMessage) error {
		return processBatch(ctx, database, encryptor, batch)
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("Waiting for messages...")

	// wait 30 seconds to sign all the messages for simplicity
	<-ctx.Done()

	log.Printf("Record Worker is finished!")
}

func processBatch(ctx context.Context, db *db.DB, crypto *crypto.KeyEncryptor, batch *messaging.BatchMessage) error {
	log.Printf("Processing batch %s with %d records", batch.BatchID, len(batch.Records))

	key, err := db.GetLeastRecentlyUsedKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get signing key: %w", err)
	}

	defer func() {
		if releaseErr := db.ReleaseKey(ctx, key.ID); releaseErr != nil {
			log.Printf("Failed to release key %d: %v", key.ID, releaseErr)
		}
	}()

	log.Printf("Using key %d to sign batch %s", key.ID, batch.BatchID)

	signatures := make(map[int][]byte, len(batch.Records))
	for _, record := range batch.Records {
		signature, err := crypto.SignPayload(key.PrivateKey, record.Payload)
		if err != nil {
			return fmt.Errorf("failed to sign record %d: %w", record.ID, err)
		}
		signatures[record.ID] = signature
	}

	if err := db.UpdateRecordSignatures(ctx, signatures, key.ID); err != nil {
		return fmt.Errorf("failed to update record signatures: %w", err)
	}

	log.Printf("Successfully processed batch %s with %d records using key %d",
		batch.BatchID, len(batch.Records), key.ID)

	return nil
}
