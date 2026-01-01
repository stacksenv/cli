package stacksenv

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

/*
Payload format (base64 encoded):
| nonce (12 bytes) | ciphertext + auth tag (16 bytes) |

The encryption uses AES-256-GCM with:
- Key derivation: SHA-256 of the shared secret
- Nonce: 12 random bytes (generated per encryption)
- AAD (Additional Authenticated Data): Used for authentication
*/

// DefaultCryptoService is the default implementation of CryptoService.
type DefaultCryptoService struct{}

// NewCryptoService creates a new crypto service instance.
func NewCryptoService() CryptoService {
	return &DefaultCryptoService{}
}

// Encrypt encrypts a slice of context data for secure transmission.
//
// The encryption process:
//  1. Marshals the data to JSON
//  2. Derives a 32-byte key from the shared secret using SHA-256
//  3. Generates a random 12-byte nonce
//  4. Encrypts using AES-256-GCM with the provided AAD
//  5. Appends nonce to ciphertext and base64 encodes the result
//
// Parameters:
//   - data: The context data to encrypt
//   - sharedSecret: The secret key for encryption (must not be empty)
//   - aad: Additional Authenticated Data for integrity verification
//
// Returns the base64-encoded encrypted payload or an error if encryption fails.
func (s *DefaultCryptoService) Encrypt(
	data []ContextData[any],
	sharedSecret string,
	aad string,
) (string, error) {
	if sharedSecret == "" {
		return "", errors.New("shared secret cannot be empty")
	}

	// Marshal data to JSON
	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	// Derive 32-byte key from shared secret
	key := sha256.Sum256([]byte(sharedSecret))

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("cipher init failed: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("gcm init failed: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	// Encrypt with AAD
	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(aad))

	// Prepend nonce to ciphertext
	payload := make([]byte, 0, len(nonce)+len(ciphertext))
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)

	// Base64 encode
	return base64.StdEncoding.EncodeToString(payload), nil
}

// Decrypt decrypts an encrypted payload and returns the context data.
//
// The decryption process:
//  1. Base64 decodes the payload
//  2. Extracts the nonce (first 12 bytes)
//  3. Derives the key from the shared secret using SHA-256
//  4. Decrypts using AES-256-GCM with the provided AAD
//  5. Unmarshals the JSON to context data
//
// Parameters:
//   - encrypted: The base64-encoded encrypted payload
//   - sharedSecret: The secret key for decryption (must not be empty)
//   - aad: Additional Authenticated Data (must match the AAD used during encryption)
//
// Returns the decrypted context data or an error if decryption fails.
func (s *DefaultCryptoService) Decrypt(
	encrypted string,
	sharedSecret string,
	aad string,
) ([]ContextData[any], error) {
	var result []ContextData[any]

	if encrypted == "" {
		return nil, errors.New("encrypted payload cannot be empty")
	}
	if sharedSecret == "" {
		return nil, errors.New("shared secret cannot be empty")
	}

	// Base64 decode
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	// Derive 32-byte key from shared secret
	key := sha256.Sum256([]byte(sharedSecret))

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cipher init failed: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm init failed: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return nil, errors.New("invalid payload size: too short")
	}

	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]

	// Decrypt with AAD
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(aad))
	if err != nil {
		return nil, fmt.Errorf("decrypt/auth failed: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return result, nil
}

// Encrypt is a convenience function that uses the default crypto service.
// It's maintained for backward compatibility.
func Encrypt(data []ContextData[any], sharedSecret, aad string) (string, error) {
	crypto := NewCryptoService()
	return crypto.Encrypt(data, sharedSecret, aad)
}

// Decrypt is a convenience function that uses the default crypto service.
// It's maintained for backward compatibility.
func Decrypt(encrypted string, sharedSecret, aad string) ([]ContextData[any], error) {
	crypto := NewCryptoService()
	return crypto.Decrypt(encrypted, sharedSecret, aad)
}
