package models

import (
	"encoding/json"
	"time"
)

type RecordStatus string

const (
	RecordStatusPending RecordStatus = "PENDING"
	RecordStatusQueued  RecordStatus = "QUEUED"
	RecordStatusSigned  RecordStatus = "SIGNED"
)

type Record struct {
	ID        int             `json:"id,omitempty" gorm:"primaryKey"`
	Payload   json.RawMessage `json:"payload" gorm:"type:jsonb;not null"`
	Signature []byte          `json:"signature,omitempty" gorm:"type:bytea"`
	SignedBy  int             `json:"signed_by,omitempty" gorm:"index"`
	SignedAt  *time.Time      `json:"signed_at,omitempty"`
	Status    RecordStatus    `json:"status" gorm:"type:varchar(10);not null;default:'PENDING'"`
}

type RecordMessage struct {
	ID      int             `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

func NewRecordMessage(r *Record) RecordMessage {
	return RecordMessage{
		ID:      r.ID,
		Payload: r.Payload,
	}
}

type SigningKey struct {
	ID         int        `json:"id,omitempty" gorm:"primaryKey"`
	PublicKey  []byte     `json:"public_key" gorm:"type:bytea;not null"`
	PrivateKey []byte     `json:"private_key,omitempty" gorm:"type:bytea;not null"`
	LastUsed   *time.Time `json:"last_used,omitempty"`
	InUse      bool       `json:"in_use" gorm:"not null;default:true"`
}
