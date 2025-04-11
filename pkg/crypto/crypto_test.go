package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateEncryptionKey()
	if err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}

	encryptor, err := NewKeyEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	originalData := []byte("test data that needs to be encrypted")

	encrypted, err := encryptor.Encrypt(originalData)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, originalData) {
		t.Errorf("Decrypted data does not match original data")
	}
}
