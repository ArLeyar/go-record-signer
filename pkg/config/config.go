package config

import (
	"encoding/base64"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL         string
	KeyCount            int
	RecordCount         int
	EncryptionKeyBase64 string
}

func LoadConfig() *Config {
	cfg := &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/recordsigner?sslmode=disable"),
		KeyCount:            getEnvAsInt("KEY_COUNT", 100),
		RecordCount:         getEnvAsInt("RECORD_COUNT", 100000),
		EncryptionKeyBase64: getEnv("ENCRYPTION_KEY", ""),
	}

	return cfg
}

func (c *Config) GetEncryptionKey() ([]byte, error) {
	if c.EncryptionKeyBase64 == "" {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(c.EncryptionKeyBase64)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
