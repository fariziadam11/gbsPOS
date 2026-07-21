package qris

import (
	"fmt"
	"strings"
)

// ConvertOptions contains options for converting static to dynamic QRIS
type ConvertOptions struct {
	Amount    float64 // Transaction amount in IDR
	FeeType   string  // "fixed" or "percentage" (optional)
	FeeValue  float64 // Fee amount (optional)
}

// Convert converts a static QRIS to dynamic by injecting amount
// Returns the dynamic QRIS string
func Convert(qrisString string, amount float64) (string, error) {
	return ConvertWithFee(qrisString, ConvertOptions{Amount: amount})
}

// ConvertWithFee converts a static QRIS to dynamic with optional fee
func ConvertWithFee(qrisString string, opts ConvertOptions) (string, error) {
	// Parse the QRIS first to validate
	parsed, err := Parse(qrisString)
	if err != nil {
		return "", fmt.Errorf("failed to parse QRIS: %v", err)
	}

	// Validate it's a static QRIS
	if parsed.IsDynamic {
		return "", fmt.Errorf("QRIS is already dynamic, cannot convert")
	}

	// Validate currency is IDR (360)
	if parsed.Currency != "360" {
		return "", fmt.Errorf("only IDR currency (360) is supported, got %s", parsed.Currency)
	}

	// Validate amount
	if opts.Amount <= 0 {
		return "", fmt.Errorf("amount must be greater than 0")
	}

	// Format amount (2 decimal places)
	amountStr := fmt.Sprintf("%.2f", opts.Amount)

	// Build the new QRIS
	var builder strings.Builder

	// 1. Payload Format Indicator (00) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "00"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 2. Change Point of Initiation Method (01) from 11 to 12
	builder.WriteString("0112") // Tag "01", length 2, value "12" (dynamic)

	// 3. Merchant Account Information (26-51) - keep as is
	for _, tag := range []string{"26", "27", "28", "29", "30", "31", "32", "33", "34", "35",
		"36", "37", "38", "39", "40", "41", "42", "43", "44", "45",
		"46", "47", "48", "49", "50", "51"} {
		if tlv := findTLV(parsed.RawTLVs, tag); tlv != nil {
			builder.WriteString(tlv.Raw)
		}
	}

	// 4. Merchant Category Code (52) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "52"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 5. Transaction Currency (53) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "53"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 6. Transaction Amount (54) - NEW TAG
	amountTag := formatAmountTag(amountStr)
	builder.WriteString(amountTag)

	// 7. Tip Indicator (55) - keep if present
	if tlv := findTLV(parsed.RawTLVs, "55"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 8. Fixed Fee (56) or Percentage Fee (57) - handle fee if provided
	if opts.FeeType != "" && opts.FeeValue > 0 {
		if opts.FeeType == "fixed" {
			feeStr := fmt.Sprintf("%.2f", opts.FeeValue)
			feeTag := fmt.Sprintf("56%02d%s", len(feeStr), feeStr)
			builder.WriteString(feeTag)
		} else if opts.FeeType == "percentage" {
			feeStr := fmt.Sprintf("%.2f", opts.FeeValue)
			feeTag := fmt.Sprintf("57%02d%s", len(feeStr), feeStr)
			builder.WriteString(feeTag)
		}
	}

	// 9. Country Code (58) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "58"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 10. Merchant Name (59) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "59"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 11. Merchant City (60) - keep as is
	if tlv := findTLV(parsed.RawTLVs, "60"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 12. Additional Data (62) - keep if present
	if tlv := findTLV(parsed.RawTLVs, "62"); tlv != nil {
		builder.WriteString(tlv.Raw)
	}

	// 13. CRC16 (63) - calculate new CRC
	dataBeforeCRC := builder.String()
	crc := CRC16(dataBeforeCRC)
	builder.WriteString(fmt.Sprintf("63%02d%s", len(crc), crc))

	return builder.String(), nil
}

// formatAmountTag creates tag 54 for transaction amount
func formatAmountTag(amount string) string {
	// Remove decimal point for QRIS format? No, keep it
	// QRIS amount format: "54" + length + value
	return fmt.Sprintf("54%02d%s", len(amount), amount)
}

// findTLV finds a TLV element by tag
func findTLV(tlvs []TLV, tag string) *TLV {
	for i := range tlvs {
		if tlvs[i].Tag == tag {
			return &tlvs[i]
		}
	}
	return nil
}

// ConvertToBase64 converts a QRIS string to base64 format for QR code generation
func ConvertToBase64(qrisString string) string {
	// QRIS is already a string that can be encoded to base64
	// Some QR code readers expect base64 encoded data
	return qrisString // Return as-is, the QR library will handle encoding
}

// ParseAmount extracts amount from dynamic QRIS
func ParseAmount(qrisString string) (float64, error) {
	parsed, err := Parse(qrisString)
	if err != nil {
		return 0, err
	}

	if parsed.Amount == "" {
		return 0, fmt.Errorf("no amount found in QRIS")
	}

	var amount float64
	_, err = fmt.Sscanf(parsed.Amount, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %s", parsed.Amount)
	}

	return amount, nil
}

// GetFee calculates the fee based on fee type and value
func GetFee(amount float64, feeType string, feeValue float64) float64 {
	if feeType == "" || feeValue <= 0 {
		return 0
	}

	switch feeType {
	case "fixed":
		return feeValue
	case "percentage":
		return amount * (feeValue / 100)
	default:
		return 0
	}
}
