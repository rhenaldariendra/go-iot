package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("email or password is incorrect")
	}

	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// EncryptPayload encrypts any JSON-serializable struct using AES-GCM
func EncryptPayload(data interface{}) (string, error) {
	// Get base64-encoded key from environment
	base64Key := ReadConfigBaseServer().AESGCMKey
	if base64Key == "" {
		return "", errors.New("ENCRYPTION_KEY not set in environment")
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return "", fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("json marshal failed: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cipher creation failed: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCM creation failed: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	// Encrypt and seal
	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)
	return hex.EncodeToString(ciphertext), nil
}

// DecryptPayload decrypts into the target struct
func DecryptPayload(encrypted string, target interface{}) error {
	// Get base64-encoded key from environment
	base64Key := ReadConfigBaseServer().AESGCMKey
	if base64Key == "" {
		return errors.New("ENCRYPTION_KEY not set in environment")
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return fmt.Errorf("invalid base64 key: %w", err)
	}
	if len(key) != 32 {
		return errors.New("encryption key must be 32 bytes")
	}

	// Decode from hex
	ciphertext, err := hex.DecodeString(encrypted)
	if err != nil {
		return fmt.Errorf("hex decode failed: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cipher creation failed: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("GCM creation failed: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(plaintext, target); err != nil {
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	return nil
}

// GenerateAES256Key generates a new 32-byte key and returns it as base64
func GenerateAES256Key() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", fmt.Errorf("key generation failed: %v", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}
