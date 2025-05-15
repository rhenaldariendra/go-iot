package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/chacha20poly1305"
	"math/big"
	rand2 "math/rand"
	"strconv"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomID(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand2.Intn(len(charset))]
	}
	return string(b)
}

func Generate32ByteKey() []byte {
	key := make([]byte, chacha20poly1305.KeySize) // 32 bytes for ChaCha20-Poly1305
	_, err := rand.Read(key)
	if err != nil {
		return nil
	}
	return key
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func CheckHash(token, hashed string) bool {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]) == hashed
}

func GenerateOTP(length int) ([]string, error) {
	otp := ""
	otpArray := make([]string, length+1) // Create a slice with length+1 to store digits and the full OTP
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10)) // Generate a random number between 0-9
		if err != nil {
			return nil, err
		}
		digit := strconv.Itoa(int(num.Int64()))
		otp += digit
		otpArray[i] = digit // Store each digit in the slice
	}
	otpArray[length] = otp // Store the full OTP as the last element
	return otpArray, nil
}
