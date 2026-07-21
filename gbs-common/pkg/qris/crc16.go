package qris

import (
	"fmt"
)

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
	return CRC16Bytes([]byte(data))
}

// CRC16Bytes calculates the CRC16-CCITT checksum for the given byte slice
func CRC16Bytes(data []byte) string {
	crc := uint16(0xFFFF) // Initial value

	for _, b := range data {
		crc = (crc << 8) ^ crc16Table[byte(crc>>8)^b]
	}

	return fmt.Sprintf("%04X", crc)
}

// CRC16WithInput calculates CRC16 with a specific initial value
func CRC16WithInput(data string, initial uint16) string {
	crc := initial

	for _, b := range []byte(data) {
		crc = (crc << 8) ^ crc16Table[byte(crc>>8)^b]
	}

	return fmt.Sprintf("%04X", crc)
}
