package db

import (
	"context"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/arleyar/go-record-signer/pkg/config"
	"github.com/arleyar/go-record-signer/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	gorm *gorm.DB
}

func New(cfg *config.Config) (*DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")
	return &DB{gorm: gormDB}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return fmt.Errorf("error getting database connection: %w", err)
	}
	return sqlDB.Close()
}

func (db *DB) CreateTables() error {
	err := db.gorm.AutoMigrate(&models.SigningKey{}, &models.Record{})
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

func (db *DB) InsertSigningKeys(keys []*models.SigningKey) error {
	if len(keys) == 0 {
		return nil
	}

	result := db.gorm.Create(keys)
	if result.Error != nil {
		return fmt.Errorf("failed to insert signing keys: %w", result.Error)
	}

	return nil
}

func (db *DB) InsertRecords(records []*models.Record) error {
	if len(records) == 0 {
		return nil
	}

	const batchSize = 1000
	totalRecords := len(records)

	result := db.gorm.CreateInBatches(records, batchSize)
	if result.Error != nil {
		return fmt.Errorf("failed to insert records: %w", result.Error)
	}

	log.Printf("Inserted %d records", totalRecords)
	return nil
}

func (db *DB) GetPendingRecords(ctx context.Context, batchSize int) ([]*models.Record, error) {
	var records []*models.Record

	result := db.gorm.WithContext(ctx).
		Where("status = ?", models.RecordStatusPending).
		Order("id").
		Limit(batchSize).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query pending records: %w", result.Error)
	}

	return records, nil
}

func (db *DB) UpdateRecordsToQueued(ctx context.Context, records []*models.Record) error {
	if len(records) == 0 {
		return nil
	}

	ids := make([]int, len(records))
	for i, record := range records {
		ids[i] = record.ID
	}

	result := db.gorm.WithContext(ctx).
		Model(&models.Record{}).
		Where("id IN ?", ids).
		Where("status = ?", models.RecordStatusPending).
		Update("status", models.RecordStatusQueued)

	if result.Error != nil {
		return fmt.Errorf("failed to update records to queued: %w", result.Error)
	}

	return nil
}

func (db *DB) GetLeastRecentlyUsedKey(ctx context.Context) (*models.SigningKey, error) {
	var key models.SigningKey

	err := db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where("in_use = ?", false).
			Order("last_used NULLS FIRST, id").
			Limit(1).
			Find(&key)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("no available signing keys found")
		}

		now := time.Now()
		result = tx.Model(&key).
			Updates(map[string]interface{}{
				"in_use":    true,
				"last_used": now,
			})

		if result.Error != nil {
			return result.Error
		}

		key.LastUsed = &now
		key.InUse = true

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get least recently used key: %w", err)
	}

	return &key, nil
}

func (db *DB) ReleaseKey(ctx context.Context, keyID int) error {
	result := db.gorm.WithContext(ctx).
		Model(&models.SigningKey{}).
		Where("id = ?", keyID).
		Update("in_use", false)

	if result.Error != nil {
		return fmt.Errorf("failed to release key %d: %w", keyID, result.Error)
	}

	return nil
}

func (db *DB) UpdateRecordSignatures(ctx context.Context, signatures map[int][]byte, keyID int) error {
	if len(signatures) == 0 {
		return nil
	}

	now := time.Now()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	caseBuilder := sq.Case("id")
	ids := make([]interface{}, 0, len(signatures))

	for id, sig := range signatures {
		caseBuilder = caseBuilder.When(sq.Expr("?", id), sq.Expr("?", sig))
		ids = append(ids, id)
	}

	query, args, err := psql.Update("records").
		Set("signature", caseBuilder).
		Set("signed_by", keyID).
		Set("signed_at", now).
		Set("status", models.RecordStatusSigned).
		Where(sq.Expr("id IN ("+sq.Placeholders(len(ids))+")", ids...)).
		Where(sq.Eq{"status": models.RecordStatusQueued}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	result := db.gorm.WithContext(ctx).Exec(query, args...)
	if result.Error != nil {
		return fmt.Errorf("failed to update signatures: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no records were updated")
	}

	return nil
}
