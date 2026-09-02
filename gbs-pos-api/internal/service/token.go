package service

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken returns a cryptographically random hex string (32 hex chars / 128 bits).
func GenerateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
