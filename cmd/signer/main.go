package main

import (
	"log"

	"github.com/arleyar/go-record-signer/internal/db"
	"github.com/arleyar/go-record-signer/pkg/config"
)

func main() {
	log.Println("Starting Record Signer")

	cfg := config.LoadConfig()

	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Signer database connection successful")
}
