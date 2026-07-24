package qris

import (
	"fmt"
)

// CRCAlgorithm defines the CRC algorithm variant
type CRCAlgorithm string

const (
	// CRCAlgorithmDefault is CRC-16-CCITT-FALSE (standard EMV QRIS)
	CRCAlgorithmDefault CRCAlgorithm = "ccitt-false"
	// CRCAlgorithmX25 is CRC-16-CCITT-KERMIT (reflected)
	CRCAlgorithmX25 CRCAlgorithm = "ccitt-kermit"
	// CRCAlgorithmANSI is CRC-16-ANSI (similar to Modbus)
	CRCAlgorithmANSI CRCAlgorithm = "ansi"
)

// CRCAlgorithmOverride allows using a custom CRC algorithm for QRIS generation
// Set this before calling ConvertWithFee if your provider uses a non-standard CRC
var CRCAlgorithmOverride CRCAlgorithm = CRCAlgorithmDefault

// CRC16-CCITT calculation for QRIS
// Polynomial: 0x1021 (x^16 + x^12 + x^5 + 1)
// Initial value: 0xFFFF

// crc16Table is the precomputed CRC16-CCITT table
var crc16Table [256]uint16

func init() {
	// Initialize CRC16 table with polynomial 0x1021
	for i := 0; i < 256; i++ {
		crc := uint16(i << 8)
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
		crc16Table[i] = crc
	}
}

// CRC16 calculates the CRC16-CCITT checksum for the given data
func CRC16(data string) string {
	return CRC16WithAlgorithm(data, CRCAlgorithmOverride)
}

// CRC16WithAlgorithm calculates CRC16 with a specific algorithm
func CRC16WithAlgorithm(data string, algorithm CRCAlgorithm) string {
	switch algorithm {
	case CRCAlgorithmX25:
		return crc16Kermit(data)
	case CRCAlgorithmANSI:
		return crc16ANSI(data)
	default:
		return crc16CCITTFalse(data)
	}
}

// crc16CCITTFalse calculates CRC-16-CCITT-FALSE
// Polynomial: 0x1021, Initial: 0xFFFF, No XOR out
func crc16CCITTFalse(data string) string {
	crc := uint16(0xFFFF)
	for _, b := range []byte(data) {
		crc = (crc << 8) ^ crc16Table[byte(crc>>8)^b]
	}
	return fmt.Sprintf("%04X", crc)
}

// crc16Kermit calculates CRC-16-CCITT-KERMIT (reflected)
// Polynomial: 0x1021, Initial: 0x0000, reflected algorithm
func crc16Kermit(data string) string {
	table := make([]uint16, 256)
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}

	crc := uint16(0x0000)
	for _, b := range []byte(data) {
		crc = (crc >> 8) ^ table[byte(crc)^b]
	}
	return fmt.Sprintf("%04X", crc)
}

// crc16ANSI calculates CRC-16-ANSI
// Polynomial: 0x8005, Initial: 0xFFFF, reflected algorithm
func crc16ANSI(data string) string {
	table := make([]uint16, 256)
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
		table[i] = crc
	}

	crc := uint16(0xFFFF)
	for _, b := range []byte(data) {
		crc = (crc >> 8) ^ table[byte(crc)^b]
	}
	return fmt.Sprintf("%04X", crc)
}

// CRC16Bytes calculates the CRC16-CCITT checksum for the given byte slice
func CRC16Bytes(data []byte) string {
	return CRC16(string(data))
}

// CRC16WithInput calculates CRC16 with a specific initial value
func CRC16WithInput(data string, initial uint16) string {
	crc := initial

	for _, b := range []byte(data) {
		crc = (crc << 8) ^ crc16Table[byte(crc>>8)^b]
	}

	return fmt.Sprintf("%04X", crc)
}
