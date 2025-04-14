package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/arleyar/go-record-signer/pkg/models"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	StreamName = "records"
	Subject    = "record.batches"
	MaxAge     = 24 * time.Hour
)

type BatchMessage struct {
	BatchID   string                 `json:"batch_id"`
	Records   []models.RecordMessage `json:"records"`
	CreatedAt time.Time              `json:"created_at"`
}

type NATSClient struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func New(url string) (*NATSClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	_, err = js.StreamInfo(StreamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     StreamName,
			Subjects: []string{Subject},
			Storage:  nats.FileStorage,
			MaxAge:   MaxAge,
		})
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create stream %s: %w", StreamName, err)
		}
	}

	return &NATSClient{
		conn: conn,
		js:   js,
	}, nil
}

func (c *NATSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *NATSClient) PublishBatch(records []*models.Record) error {
	recordMessages := make([]models.RecordMessage, len(records))
	for i, record := range records {
		recordMessages[i] = models.NewRecordMessage(record)
	}

	msg := BatchMessage{
		BatchID:   uuid.New().String(),
		Records:   recordMessages,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal batch message: %w", err)
	}

	_, err = c.js.Publish(Subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish batch message: %w", err)
	}

	return nil
}
