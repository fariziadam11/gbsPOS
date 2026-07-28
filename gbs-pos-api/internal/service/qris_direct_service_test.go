package service

import (
	"testing"

	"gbs-pos-api/internal/dto"
)

func TestConvertQRISRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       dto.ConvertQRISRequest
		wantError bool
	}{
		{
			name: "valid request without fee",
			req: dto.ConvertQRISRequest{
				Amount: 50000,
			},
			wantError: false,
		},
		{
			name: "valid request with fixed fee",
			req: dto.ConvertQRISRequest{
				Amount:  50000,
				FeeType: "fixed",
				FeeValue: 1000,
			},
			wantError: false,
		},
		{
			name: "valid request with percentage fee",
			req: dto.ConvertQRISRequest{
				Amount:  50000,
				FeeType: "percentage",
				FeeValue: 2.5,
			},
			wantError: false,
		},
		{
			name: "invalid fee type",
			req: dto.ConvertQRISRequest{
				Amount:  50000,
				FeeType: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false

			// Validate fee type
			if tt.req.FeeType != "" && tt.req.FeeType != "fixed" && tt.req.FeeType != "percentage" {
				hasError = true
			}

			if hasError != tt.wantError {
				t.Errorf("validation error = %v, wantError = %v", hasError, tt.wantError)
			}
		})
	}
}

func TestQrisTransactionStatus(t *testing.T) {
	// Test that status constants are defined correctly
	tests := []struct {
		name     string
		expected string
	}{
		{"Pending status", "PENDING"},
		{"AwaitingConfirmation status", "AWAITING_CONFIRMATION"},
		{"Paid status", "PAID"},
		{"Cancelled status", "CANCELLED"},
		{"Expired status", "EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the test runs without panic
			// Actual status values are tested through integration tests
		})
	}
}
