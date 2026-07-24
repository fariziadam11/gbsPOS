package qris

import (
	"fmt"
	"strings"
)

// SkipCRCValidation allows parsing QRIS strings with invalid CRC
// Useful for development or QRIS strings generated with different CRC algorithms
var SkipCRCValidation = true // Default to true to allow proprietary CRC from providers like Gojek

// TLV represents a Tag-Length-Value element
type TLV struct {
	Tag     string
	Name    string
	Length  int
	Value   string
	Raw     string
}

// MerchantAccountInfo represents merchant account information
type MerchantAccountInfo struct {
	ProviderID    string // NPS (Penyedia Jasa Pembayaran)
	MerchantID    string
	TerminalID    string
	InstitutionCode string
}

// ParsedQRIS represents the parsed QRIS data
type ParsedQRIS struct {
	Version            string                 // Payload Format Indicator (tag 00)
	InitiateMethod     string                 // Point of Initiation Method (tag 01): "11" = static, "12" = dynamic
	MerchantAccounts   []MerchantAccountInfo  // Merchant Account Information (tag 26-52)
	MCC               string                 // Merchant Category Code (tag 52)
	Currency          string                 // Transaction Currency (tag 53): "360" = IDR
	Amount            string                 // Transaction Amount (tag 54)
	TipIndicator      string                 // Tip Indicator (tag 55)
	FixedFee          string                 // Fixed Fee (tag 56)
	PercentageFee     string                 // Percentage Fee (tag 57)
	CountryCode       string                 // Country Code (tag 58)
	MerchantName      string                 // Merchant Name (tag 59)
	MerchantCity      string                 // Merchant City (tag 60)
	PostalCode        string                 // Postal Code (tag 61)
	AdditionalData    string                 // Additional Data (tag 62)
	CRC               string                 // CRC (tag 63)
	RawTLVs           []TLV                 // All raw TLV elements
	IsStatic          bool                   // True if static QRIS
	IsDynamic         bool                   // True if dynamic QRIS
}

// QRIS Tag definitions
var tagNames = map[string]string{
	"00": "Payload Format Indicator",
	"01": "Point of Initiation Method",
	"02": "Acquirer System Trace Audit Number",
	"03": "Acquirer Business Code",
	"05": "Forwarding Business Code",
	"06": "Amount, Transaction Fee",
	"07": "Amount, Settlement Fee",
	"08": "Amount, Cardholder Billing Fee",
	"09": "Settlement Analysis Code",
	"10": "Amount, Transaction",
	"11": "Amount, Transaction Fee",
	"12": "Amount, Settlement Fee",
	"13": "Amount, Cardholder Billing Fee",
	"14": "Amount, Cardholder Billing Fee",
	"15": "Settlement Analysis Code",
	"16": "Amount, Transaction",
	"17": "Amount, Transaction Fee",
	"18": "Amount, Settlement Fee",
	"19": "Amount, Cardholder Billing Fee",
	"20": "Amount, Transaction",
	"21": "Amount, Transaction Fee",
	"22": "Amount, Settlement Fee",
	"23": "Amount, Cardholder Billing Fee",
	"24": "Settlement Analysis Code",
	"25": "Amount, Transaction",
	"26": "Merchant Account Information",
	"27": "Merchant Account Information",
	"28": "Merchant Account Information",
	"29": "Merchant Account Information",
	"30": "Merchant Account Information",
	"31": "Merchant Account Information",
	"32": "Merchant Account Information",
	"33": "Merchant Account Information",
	"34": "Merchant Account Information",
	"35": "Merchant Account Information",
	"36": "Merchant Account Information",
	"37": "Merchant Account Information",
	"38": "Merchant Account Information",
	"39": "Merchant Account Information",
	"40": "Merchant Account Information",
	"41": "Merchant Account Information",
	"42": "Merchant Account Information",
	"43": "Merchant Account Information",
	"44": "Merchant Account Information",
	"45": "Merchant Account Information",
	"46": "Merchant Account Information",
	"47": "Merchant Account Information",
	"48": "Merchant Account Information",
	"49": "Merchant Account Information",
	"50": "Merchant Account Information",
	"51": "Merchant Account Information",
	"52": "Merchant Category Code",
	"53": "Transaction Currency",
	"54": "Transaction Amount",
	"55": "Tip Indicator",
	"56": "Tip Fixed",
	"57": "Tip Percentage",
	"58": "Country Code",
	"59": "Merchant Name",
	"60": "Merchant City",
	"61": "Merchant Postal Code",
	"62": "Additional Data",
	"63": "CRC16",
}

// Merchant Account Info sub-tags
var merchantAccountSubTags = map[string]string{
	"00": "Global Unique Identifier",
	"01": "Payment Network Association ID",
	"02": "Acquirer ID",
	"03": "Acquirer Business Code",
	"08": "Reserved for PDS",
	"09": "Reserved for PDS",
	"10": "Reserved for PDS",
	"11": "Reserved for PDS",
	"12": "Reserved for PDS",
	"13": "Reserved for PDS",
	"14": "Reserved for PDS",
	"15": "Reserved for PDS",
	"16": "Reserved for PDS",
	"17": "Reserved for PDS",
	"18": "Reserved for PDS",
	"19": "Reserved for PDS",
	"20": "Reserved for PDS",
	"21": "Reserved for PDS",
	"22": "Reserved for PDS",
	"23": "Reserved for PDS",
	"24": "Reserved for PDS",
	"25": "Reserved for PDS",
	"26": "Reserved for PDS",
	"27": "Reserved for PDS",
	"28": "Reserved for PDS",
	"29": "Reserved for PDS",
	"30": "Reserved for PDS",
	"31": "Reserved for PDS",
	"32": "Reserved for PDS",
	"33": "Reserved for PDS",
	"34": "Reserved for PDS",
	"35": "Reserved for PDS",
	"36": "Reserved for PDS",
	"37": "Reserved for PDS",
	"38": "Reserved for PDS",
	"39": "Reserved for PDS",
	"40": "Reserved for PDS",
	"41": "Reserved for PDS",
	"42": "Reserved for PDS",
	"43": "Reserved for PDS",
	"44": "Reserved for PDS",
	"45": "Reserved for PDS",
	"46": "Reserved for PDS",
	"47": "Reserved for PDS",
	"48": "Reserved for PDS",
	"49": "Reserved for PDS",
	"50": "Reserved for PDS",
	"51": "Reserved for PDS",
	"52": "Reserved for PDS",
	"53": "Reserved for PDS",
	"54": "Reserved for PDS",
	"55": "Reserved for PDS",
	"56": "Reserved for PDS",
	"57": "Reserved for PDS",
	"58": "Reserved for PDS",
	"59": "Reserved for PDS",
	"60": "Reserved for PDS",
	"61": "Reserved for PDS",
	"62": "Reserved for PDS",
	"63": "Reserved for PDS",
	"64": "Reserved for PDS",
	"65": "Reserved for PDS",
	"66": "Reserved for PDS",
	"67": "Reserved for PDS",
	"68": "Reserved for PDS",
	"69": "Reserved for PDS",
	"70": "Reserved for PDS",
	"71": "Reserved for PDS",
	"72": "Reserved for PDS",
	"73": "Reserved for PDS",
	"74": "Reserved for PDS",
	"75": "Reserved for PDS",
	"76": "Reserved for PDS",
	"77": "Reserved for PDS",
	"78": "Reserved for PDS",
	"79": "Reserved for PDS",
	"80": "Reserved for PDS",
	"81": "Reserved for PDS",
	"82": "Reserved for PDS",
	"83": "Reserved for PDS",
	"84": "Reserved for PDS",
	"85": "Reserved for PDS",
	"86": "Reserved for PDS",
	"87": "Reserved for PDS",
	"88": "Reserved for PDS",
	"89": "Reserved for PDS",
	"90": "Reserved for PDS",
	"91": "Acquirer Transaction ID",
	"92": "Acquirer Reference Number",
	"93": "Reserved for PDS",
	"94": "Reserved for PDS",
	"95": "Reserved for PDS",
	"96": "Reserved for PDS",
	"97": "Reserved for PDS",
	"98": "Payee Invoice",
	"99": "Reserved for PDS",
}

// ParseTLV parses a QRIS string into TLV elements
func ParseTLV(qrisString string) ([]TLV, error) {
	var tlvs []TLV
	data := strings.TrimSpace(qrisString)

	if data == "" {
		return nil, fmt.Errorf("empty QRIS string")
	}

	// Check if starts with standard indicator
	if !strings.HasPrefix(data, "00") {
		return nil, fmt.Errorf("invalid QRIS: must start with '00' (Payload Format Indicator)")
	}

	i := 0
	for i < len(data) {
		if i+2 > len(data) {
			return nil, fmt.Errorf("incomplete data at position %d: missing tag", i)
		}

		tag := data[i : i+2]
		i += 2

		if i+2 > len(data) {
			return nil, fmt.Errorf("incomplete data at position %d: missing length for tag %s", i, tag)
		}

		lengthStr := data[i : i+2]
		i += 2

		length := 0
		fmt.Sscanf(lengthStr, "%02d", &length)

		if i+length > len(data) {
			return nil, fmt.Errorf("incomplete data at position %d: tag %s expects length %d but only %d bytes available",
				i, tag, length, len(data)-i)
		}

		value := data[i : i+length]
		i += length

		name := tagNames[tag]
		if name == "" {
			name = "Unknown"
		}

		tlvs = append(tlvs, TLV{
			Tag:    tag,
			Name:   name,
			Length: length,
			Value:  value,
			Raw:    tag + lengthStr + value,
		})
	}

	return tlvs, nil
}

// Parse parses a QRIS string into a structured object
// Set SkipCRCValidation to true if the QRIS string has an invalid CRC
func Parse(qrisString string) (*ParsedQRIS, error) {
	tlvs, err := ParseTLV(qrisString)
	if err != nil {
		return nil, err
	}

	result := &ParsedQRIS{
		RawTLVs: tlvs,
	}

	// Validate CRC first (skip if QRIS has non-standard CRC)
	if SkipCRCValidation {
		// Skip CRC validation
	} else if err := ValidateCRC(qrisString); err != nil {
		return nil, fmt.Errorf("CRC validation failed: %v", err)
	}

	// Extract known tags
	for _, tlv := range tlvs {
		switch tlv.Tag {
		case "00":
			result.Version = tlv.Value
		case "01":
			result.InitiateMethod = tlv.Value
			result.IsStatic = tlv.Value == "11"
			result.IsDynamic = tlv.Value == "12"
		case "52":
			result.MCC = tlv.Value
		case "53":
			result.Currency = tlv.Value
		case "54":
			result.Amount = tlv.Value
		case "55":
			result.TipIndicator = tlv.Value
		case "56":
			result.FixedFee = tlv.Value
		case "57":
			result.PercentageFee = tlv.Value
		case "58":
			result.CountryCode = tlv.Value
		case "59":
			result.MerchantName = tlv.Value
		case "60":
			result.MerchantCity = tlv.Value
		case "61":
			result.PostalCode = tlv.Value
		case "63":
			result.CRC = tlv.Value
		case "26", "27", "28", "29", "30", "31", "32", "33", "34", "35",
			"36", "37", "38", "39", "40", "41", "42", "43", "44", "45",
			"46", "47", "48", "49", "50", "51":
			// Parse merchant account info sub-tags
			accountInfo := parseMerchantAccountInfo(tlv.Value)
			accountInfo.ProviderID = detectProviderFromAccount(tlv.Value)
			result.MerchantAccounts = append(result.MerchantAccounts, accountInfo)
		}
	}

	// If initiate method is empty, default to static
	if result.InitiateMethod == "" {
		result.InitiateMethod = "11"
		result.IsStatic = true
	}

	return result, nil
}

// parseMerchantAccountInfo parses merchant account information sub-tags
func parseMerchantAccountInfo(value string) MerchantAccountInfo {
	info := MerchantAccountInfo{}
	i := 0

	for i < len(value) {
		if i+2 > len(value) {
			break
		}

		subTag := value[i : i+2]
		i += 2

		if i+2 > len(value) {
			break
		}

		subLengthStr := value[i : i+2]
		i += 2

		subLength := 0
		fmt.Sscanf(subLengthStr, "%02d", &subLength)

		if i+subLength > len(value) {
			break
		}

		subValue := value[i : i+subLength]
		i += subLength

		switch subTag {
		case "00":
			// Global Unique Identifier - usually contains merchant info
			info.MerchantID = subValue
		case "01":
			// Payment Network Association ID (NPA/PJP code)
			info.ProviderID = subValue
		case "02":
			info.InstitutionCode = subValue
		}
	}

	return info
}

// detectProviderFromAccount detects payment provider from merchant account info
func detectProviderFromAccount(value string) string {
	// Common provider prefixes in QRIS
	providers := map[string]string{
		"DANA":     "DANA",
		"OVO":      "OVO",
		"SHOPEEPAY": "ShopeePay",
		"GOPAY":    "GoPay",
		"LINK":     "LinkAja",
		"SAKUKU":   "Sakuku",
		"MANDIRI":  "Mandiri",
		"BNI":      "BNI",
		"BRI":      "BRI",
		"BTPN":     "BTPN",
		"BCA":      "BCA",
		"CIMB":     "CIMB",
		"PERMATA":  "Permata",
	}

	upperValue := strings.ToUpper(value)
	for prefix, name := range providers {
		if strings.Contains(upperValue, prefix) {
			return name
		}
	}

	// Try to extract from sub-tag 01 (NPA/PJP code)
	info := parseMerchantAccountInfo(value)
	if info.ProviderID != "" {
		return info.ProviderID
	}

	return "Unknown"
}

// ValidateCRC validates the CRC16 checksum of a QRIS string
func ValidateCRC(qrisString string) error {
	data := strings.TrimSpace(qrisString)

	// Find the CRC tag (63)
	crcIndex := strings.LastIndex(data, "63")
	if crcIndex == -1 {
		return fmt.Errorf("CRC tag (63) not found")
	}

	// Extract the data before CRC and the CRC value
	// Format: ...63[LL][CRC_VALUE]
	if crcIndex+4 > len(data) {
		return fmt.Errorf("incomplete CRC data")
	}

	lengthStr := data[crcIndex+2 : crcIndex+4]
	length := 0
	fmt.Sscanf(lengthStr, "%02d", &length)

	if crcIndex+4+length > len(data) {
		return fmt.Errorf("incomplete CRC value")
	}

	expectedCRC := data[crcIndex+4 : crcIndex+4+length]

	// Calculate CRC from data before CRC tag, INCLUDING "63" tag and "LL" length indicator
	// Per EMVCo spec and standard QRIS implementations, CRC is calculated over
	// all data including the CRC tag header (but not the CRC value itself)
	dataBeforeCRC := data[:crcIndex] + "63" + lengthStr
	calculatedCRC := CRC16(dataBeforeCRC)

	if strings.ToUpper(expectedCRC) != strings.ToUpper(calculatedCRC) {
		return fmt.Errorf("CRC mismatch: expected %s, got %s", expectedCRC, calculatedCRC)
	}

	return nil
}

// IsStatic checks if the QRIS is static
func IsStatic(qrisString string) bool {
	parsed, err := Parse(qrisString)
	if err != nil {
		return false
	}
	return parsed.IsStatic
}

// IsDynamic checks if the QRIS is dynamic
func IsDynamic(qrisString string) bool {
	parsed, err := Parse(qrisString)
	if err != nil {
		return false
	}
	return parsed.IsDynamic
}
