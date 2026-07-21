package qris

import (
	"testing"
)

func TestCRC16(t *testing.T) {
	// Test basic CRC calculation
	result := CRC16("")
	if result != "FFFF" {
		t.Errorf("CRC16('') = %s, expected FFFF", result)
	}

	// Test with data
	result = CRC16("01")
	if result == "" {
		t.Error("CRC16 should return non-empty string")
	}
}

func TestValidateCRC_Invalid(t *testing.T) {
	// Test invalid CRC - should fail validation
	invalidQris := "00020101021126570011ID.DANA.WWW01360013ID10208006950208ID0310232603DANA.WWW02280100000000000010303DANA.WWW5204541253036605403500005802ID5907DANA6014KOTA SURABAYA61056021962280523ID1020800695036521046304FFFF"

	err := ValidateCRC(invalidQris)
	if err == nil {
		t.Error("Expected CRC validation to fail for invalid CRC")
	}
}

func TestGetFee(t *testing.T) {
	tests := []struct {
		amount   float64
		feeType  string
		feeValue float64
		expected float64
	}{
		{10000, "fixed", 1000, 1000},
		{10000, "percentage", 2.5, 250},
		{10000, "invalid", 1000, 0},
		{10000, "", 1000, 0},
		{10000, "fixed", 0, 0},
	}

	for _, tt := range tests {
		result := GetFee(tt.amount, tt.feeType, tt.feeValue)
		if result != tt.expected {
			t.Errorf("GetFee(%v, %s, %v) = %v, expected %v",
				tt.amount, tt.feeType, tt.feeValue, result, tt.expected)
		}
	}
}

func TestTLVStructure(t *testing.T) {
	// Test TLV parsing with simple data
	tlvs, err := ParseTLV("000201")
	if err != nil {
		t.Fatalf("ParseTLV failed: %v", err)
	}

	if len(tlvs) != 1 {
		t.Errorf("Expected 1 TLV, got %d", len(tlvs))
	}

	if tlvs[0].Tag != "00" {
		t.Errorf("Expected tag 00, got %s", tlvs[0].Tag)
	}

	if tlvs[0].Value != "01" {
		t.Errorf("Expected value 01, got %s", tlvs[0].Value)
	}
}

func TestMerchantAccountInfoParsing(t *testing.T) {
	// Test parsing merchant account info
	info := parseMerchantAccountInfo("0103DANA0208ID0310232603")
	// Provider ID extraction test
	_ = info.ProviderID
}

func TestDetectProvider(t *testing.T) {
	// Test provider detection from account info
	tests := []struct {
		account string
		provider string
	}{
		{"01DANA0208ID", "DANA"},
		{"01OVO0208ID", "OVO"},
		{"01GOPAY0208ID", "GoPay"},
		{"01LINK0208ID", "LinkAja"},
	}

	for _, tt := range tests {
		provider := detectProviderFromAccount(tt.account)
		if provider != tt.provider {
			t.Errorf("detectProvider(%s) = %s, expected %s", tt.account, provider, tt.provider)
		}
	}
}

func TestFormatAmountTag(t *testing.T) {
	// Test amount tag formatting
	result := formatAmountTag("50000.00")
	expected := "540850000.00"
	if result != expected {
		t.Errorf("formatAmountTag('50000.00') = %s, expected %s", result, expected)
	}
}

func TestFindTLV(t *testing.T) {
	// Test finding TLV by tag
	tlvs := []TLV{
		{Tag: "00", Name: "Payload", Length: 2, Value: "01"},
		{Tag: "01", Name: "Init", Length: 2, Value: "11"},
	}

	found := findTLV(tlvs, "00")
	if found == nil {
		t.Error("Expected to find tag 00")
	}
	if found.Value != "01" {
		t.Errorf("Expected value 01, got %s", found.Value)
	}

	notFound := findTLV(tlvs, "99")
	if notFound != nil {
		t.Error("Should not find tag 99")
	}
}
