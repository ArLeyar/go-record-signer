package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arleyar/go-record-signer/internal/db"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/crypto"
	"github.com/arleyar/go-record-signer/pkg/models"
	"github.com/google/uuid"
)

func main() {
	start := time.Now()
	log.Println("Starting DB initialization")

	cfg := config.LoadConfig()

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Creating database schema...")
	if err := database.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	encryptionKey, err := cfg.GetEncryptionKey()
	if err != nil {
		log.Fatalf("Failed to decode encryption key: %v", err)
	}

	if encryptionKey == nil {
		encryptionKey, err = crypto.GenerateEncryptionKey()
		if err != nil {
			log.Fatalf("Failed to generate encryption key: %v", err)
		}

		encodedKey := base64.StdEncoding.EncodeToString(encryptionKey)
		log.Printf("Generated new encryption key: %s", encodedKey)
		log.Printf("IMPORTANT: Save this key in your ENCRYPTION_KEY environment variable for future use")
	}

	encryptor, err := crypto.NewKeyEncryptor(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create key encryptor: %v", err)
	}

	log.Printf("Generating %d Ed25519 keys...", cfg.KeyCount)
	keys, err := generateKeys(cfg.KeyCount, encryptor)
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	log.Println("Storing encrypted keys in database...")
	if err := database.InsertSigningKeys(keys); err != nil {
		log.Fatalf("Failed to store keys in database: %v", err)
	}
	log.Printf("Generated and stored %d keys", cfg.KeyCount)

	log.Printf("Generating %d records...", cfg.RecordCount)
	records, err := generateRecords(cfg.RecordCount)
	if err != nil {
		log.Fatalf("Failed to generate records: %v", err)
	}

	log.Println("Storing records in database...")
	if err := database.InsertRecords(records); err != nil {
		log.Fatalf("Failed to store records: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("DB initialization complete in %v", elapsed)
	log.Printf("%d encrypted keys generated and stored", cfg.KeyCount)
	log.Printf("%d unsigned records created", cfg.RecordCount)

	os.Exit(0)
}

func generateKey(encryptor *crypto.KeyEncryptor) (*models.SigningKey, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	encryptedKey, err := encryptor.Encrypt(pkcs8Key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	return &models.SigningKey{
		PublicKey:  pubKey,
		PrivateKey: encryptedKey,
		InUse:      false,
	}, nil
}

func generateKeys(count int, encryptor *crypto.KeyEncryptor) ([]*models.SigningKey, error) {
	keys := make([]*models.SigningKey, 0, count)
	for i := 0; i < count; i++ {
		key, err := generateKey(encryptor)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key %d: %w", i+1, err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func generateRecord() (*models.Record, error) {
	payload := map[string]interface{}{
		"id":        uuid.NewString(),
		"timestamp": time.Now().UnixNano(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	record := &models.Record{
		Payload: json.RawMessage(jsonData),
	}

	return record, nil
}

func generateRecords(count int) ([]*models.Record, error) {
	records := make([]*models.Record, 0, count)

	for i := 0; i < count; i++ {
		record, err := generateRecord()
		if err != nil {
			return nil, fmt.Errorf("failed to generate record %d: %w", i, err)
		}

		records = append(records, record)
	}

	return records, nil
}
