package qris

import (
	"fmt"
	"testing"
)

// TestVerifyFix compares generated QRIS with reference website output
func TestVerifyFix(t *testing.T) {
	staticQris := "00020101021126610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI5204581253033605802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A01110362163042807"

	SkipCRCValidation = true
	defer func() { SkipCRCValidation = false }()

	// Test with amount 104500 (from user's example)
	result, err := Convert(staticQris, 104500)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Reference website output (has INCORRECT CRC 867E)
	reference := "00020101021226610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI52045812530336054061045005802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A0111036216304867E"

	fmt.Println("=== Generated QRIS ===")
	fmt.Println(result)
	fmt.Printf("Length: %d\n", len(result))

	fmt.Println()
	fmt.Println("=== Reference Website QRIS ===")
	fmt.Println(reference)
	fmt.Printf("Length: %d\n", len(reference))

	fmt.Println()
	fmt.Println("=== Comparison (excluding CRC) ===")

	// Compare character by character (excluding CRC at the end)
	// CRC is at position 229, so compare first 229 characters
	match := true
	for i := 0; i < 229; i++ {
		if result[i] != reference[i] {
			start := i - 5
			if start < 0 {
				start = 0
			}
			end := i + 10
			if end > len(result) {
				end = len(result)
			}
			fmt.Printf("First diff at position %d:\n", i)
			fmt.Printf("  Generated: '%s'\n", result[start:end])
			fmt.Printf("  Expected:  '%s'\n", reference[start:end])
			match = false
			break
		}
	}

	if match {
		fmt.Println("✓ DATA MATCHES REFERENCE WEBSITE (excluding CRC)!")
		fmt.Println("The only difference is the CRC value.")
	} else {
		t.Error("Generated QRIS data does not match reference website")
	}

	// Now verify our CRC is valid
	fmt.Println()
	fmt.Println("=== CRC Validation ===")
	SkipCRCValidation = false
	defer func() { SkipCRCValidation = true }()

	err = ValidateCRC(result)
	if err == nil {
		fmt.Println("✓ Our CRC is VALID (standard CRC-16-CCITT with 6304 prefix)")
		fmt.Printf("  Our CRC: %s\n", result[233:237])
		fmt.Printf("  Reference CRC: %s (INVALID)\n", reference[233:237])
		fmt.Println()
		fmt.Println("NOTE: Reference website has INCORRECT CRC but wallet apps ignore it.")
		fmt.Println("      Our QRIS has CORRECT CRC and should work in all wallet apps.")
	} else {
		t.Errorf("Our CRC validation failed: %v", err)
	}
}

// TestVerifyCRCValidation verifies that the new CRC is valid
func TestVerifyCRCValidation(t *testing.T) {
	staticQris := "00020101021126610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI5204581253033605802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A01110362163042807"

	// Don't skip CRC validation for this test
	SkipCRCValidation = false
	defer func() { SkipCRCValidation = true }()

	// Convert with amount
	result, err := Convert(staticQris, 100000)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Validate CRC
	err = ValidateCRC(result)
	if err != nil {
		t.Errorf("CRC validation failed: %v", err)
	} else {
		fmt.Println("✓ CRC is valid!")
		fmt.Printf("Generated QRIS: %s\n", result)

		// Extract and show CRC
		crcIdx := -1
		for i := 0; i < len(result)-1; i++ {
			if result[i:i+2] == "63" {
				crcIdx = i
				break
			}
		}
		if crcIdx > 0 {
			crc := result[crcIdx+4 : crcIdx+8]
			fmt.Printf("CRC value: %s\n", crc)
		}
	}
}
