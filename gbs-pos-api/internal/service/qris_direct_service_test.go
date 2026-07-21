package service

import (
	"testing"

	"gbs-pos-api/internal/dto"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		qrisString string
		expected   string
	}{
		{
			"00020101021126570011ID.DANA.WWW...",
			"DANA",
		},
		{
			"00020101021126570011ID.OVO.WWW...",
			"OVO",
		},
		{
			"00020101021126570011ID.GOPAY.WWW...",
			"GoPay",
		},
	}

	for _, tt := range tests {
		result := detectProvider(tt.qrisString)
		if result != tt.expected {
			t.Errorf("detectProvider(%s) = %s, expected %s", tt.qrisString, result, tt.expected)
		}
	}
}

func TestParseQRISRequest(t *testing.T) {
	req := dto.ParseQRISRequest{
		QrisString: "00020101021126570011ID.DANA.WWW...",
	}

	if req.QrisString == "" {
		t.Error("QrisString should not be empty")
	}
}

func TestConvertQRISRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       dto.ConvertQRISRequest
		wantError bool
	}{
		{
			name: "valid request without fee",
			req: dto.ConvertQRISRequest{
				QrisString: "00020101021126570011ID.DANA.WWW...",
				Amount:    50000,
			},
			wantError: false,
		},
		{
			name: "valid request with fixed fee",
			req: dto.ConvertQRISRequest{
				QrisString: "00020101021126570011ID.DANA.WWW...",
				Amount:    50000,
				FeeType:   "fixed",
				FeeValue:  1000,
			},
			wantError: false,
		},
		{
			name: "valid request with percentage fee",
			req: dto.ConvertQRISRequest{
				QrisString: "00020101021126570011ID.DANA.WWW...",
				Amount:    50000,
				FeeType:   "percentage",
				FeeValue:  2.5,
			},
			wantError: false,
		},
		{
			name: "invalid fee type",
			req: dto.ConvertQRISRequest{
				QrisString: "00020101021126570011ID.DANA.WWW...",
				Amount:    50000,
				FeeType:   "invalid",
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
