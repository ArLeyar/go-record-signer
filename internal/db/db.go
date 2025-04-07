package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/arleyar/go-record-signer/pkg/config"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	conn, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")
	return &DB{conn: conn}, nil
}

func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}
