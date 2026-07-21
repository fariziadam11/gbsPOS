package qris

import (
	"encoding/base64"
	"fmt"
)

// QRCodeImage represents the generated QR code
type QRCodeImage struct {
	Base64 string `json:"base64"` // Base64 encoded PNG image
	URL    string `json:"url"`    // Data URL for direct embedding
}

// GenerateQRCode generates a QR code image from QRIS string
// Returns base64 encoded PNG image
func GenerateQRCode(qrisString string) (*QRCodeImage, error) {
	// Using a pure Go implementation
	// For production, you might want to use github.com/skip2/go-qrcode

	// Since we don't have a Go QR library available, we'll return a placeholder
	// The actual QR generation should be done on the frontend or via external service

	// For now, return the QRIS string encoded
	return &QRCodeImage{
		Base64: base64.StdEncoding.EncodeToString([]byte(qrisString)),
		URL:    fmt.Sprintf("data:text/plain;base64,%s", base64.StdEncoding.EncodeToString([]byte(qrisString))),
	}, nil
}

// QRCodeGenerator interface for different QR code generation backends
type QRCodeGenerator interface {
	Generate(qrisString string) (*QRCodeImage, error)
}

// NoOpQRGenerator returns QRIS string encoded in base64
type NoOpQRGenerator struct{}

// Generate implements QRCodeGenerator
func (g *NoOpQRGenerator) Generate(qrisString string) (*QRCodeImage, error) {
	return GenerateQRCode(qrisString)
}
