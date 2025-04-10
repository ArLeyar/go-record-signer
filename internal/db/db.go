package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/models"
	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
	sb   sq.StatementBuilderType
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
	return &DB{
		conn: conn,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}

func (db *DB) CreateTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS signing_keys (
			id SERIAL PRIMARY KEY,
			public_key BYTEA NOT NULL,
			private_key BYTEA NOT NULL,
			last_used TIMESTAMP,
			in_use BOOLEAN NOT NULL DEFAULT true
		);

		CREATE TABLE IF NOT EXISTS records (
			id SERIAL PRIMARY KEY,
			payload JSONB NOT NULL,
			signature BYTEA,
			signed_by INTEGER REFERENCES signing_keys(id),
			signed_at TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func (db *DB) InsertSigningKeys(keys []*models.SigningKey) error {
	if len(keys) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := db.sb.Insert("signing_keys").
		Columns("public_key", "private_key", "in_use")

	for _, key := range keys {
		query = query.Values(key.PublicKey, key.PrivateKey, key.InUse)
	}

	_, err := query.RunWith(db.conn).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert signing keys: %w", err)
	}

	return nil
}

func (db *DB) InsertRecords(records []*models.Record) error {
	if len(records) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	batchSize := 1000
	totalRecords := len(records)

	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		query := db.sb.Insert("records").Columns("payload")
		for j := i; j < end; j++ {
			query = query.Values(records[j].Payload)
		}

		_, err := query.RunWith(db.conn).ExecContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert records batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Inserted %d of %d records", end, totalRecords)
	}

	return nil
}
