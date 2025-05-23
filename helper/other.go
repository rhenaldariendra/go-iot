package helper

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/google/uuid"
	"strconv"
	"time"
)

func GetCurrentTime() time.Time {
	return time.Now()
}

func GenerateUID() string {
	return uuid.New().String()
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length/2) // Each byte is represented by 2 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func StringToInt(s string) (int, error) {
	if s == "" {
		return 404, nil
	}
	return strconv.Atoi(s)
}
