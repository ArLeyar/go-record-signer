package models

import (
	"encoding/json"
	"time"
)

type Record struct {
	ID        string          `json:"id" db:"id"`
	Payload   json.RawMessage `json:"payload" db:"payload"`
	Signature []byte          `json:"signature,omitempty" db:"signature"`
	SignedBy  string          `json:"signed_by,omitempty" db:"signed_by"`
	SignedAt  *time.Time      `json:"signed_at,omitempty" db:"signed_at"`
}

type SigningKey struct {
	ID        string     `json:"id" db:"id"`
	KeyLabel  string     `json:"key_label" db:"key_label"`
	PublicKey []byte     `json:"public_key" db:"public_key"`
	LastUsed  *time.Time `json:"last_used,omitempty" db:"last_used"`
	InUse     bool       `json:"in_use" db:"in_use"`
}
