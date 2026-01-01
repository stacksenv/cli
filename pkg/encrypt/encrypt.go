package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type ContextData[T any] struct {
	Property string `json:"property"`
	Value    T      `json:"value"`
}

/*
Payload format (base64):

| nonce (12 bytes) | ciphertext + auth tag |
*/

// Encrypt encrypts data for server-to-server communication
func Encrypt(
	data []ContextData[any],
	sharedSecret string,
	aad string, // e.g. "serviceA->serviceB|v1"
) (string, error) {

	if sharedSecret == "" {
		return "", errors.New("shared secret cannot be empty")
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	// Derive fixed 32-byte key (OK for server secrets)
	key := sha256.Sum256([]byte(sharedSecret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("cipher init failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm init failed: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce failed: %w", err)
	}

	ciphertext := gcm.Seal(
		nil,
		nonce,
		plaintext,
		[]byte(aad),
	)

	nonce = append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(nonce), nil
}

// Decrypt decrypts server-to-server encrypted data
func Decrypt(
	encrypted string,
	sharedSecret string,
	aad string,
) ([]ContextData[any], error) {

	var result []ContextData[any]

	if encrypted == "" {
		return nil, errors.New("encrypted payload empty")
	}
	if sharedSecret == "" {
		return nil, errors.New("shared secret empty")
	}

	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	key := sha256.Sum256([]byte(sharedSecret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cipher init failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm init failed: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, errors.New("invalid payload size")
	}

	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]

	plaintext, err := gcm.Open(
		nil,
		nonce,
		ciphertext,
		[]byte(aad),
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt/auth failed: %w", err)
	}

	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return result, nil
}
