package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
)

type KeyEncryptor struct {
	encryptionKey []byte
}

func NewKeyEncryptor(encryptionKey []byte) (*KeyEncryptor, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes (256 bits)")
	}

	return &KeyEncryptor{
		encryptionKey: encryptionKey,
	}, nil
}

func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

func (e *KeyEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (e *KeyEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

func (e *KeyEncryptor) SignPayload(privateKeyBytes []byte, payload []byte) ([]byte, error) {
	decryptedKey, err := e.Decrypt(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(decryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	if key, ok := privateKey.(ed25519.PrivateKey); ok {
		return ed25519.Sign(key, payload), nil
	} else {
		return nil, fmt.Errorf("unsupported private key type: %T", privateKey)
	}
}
