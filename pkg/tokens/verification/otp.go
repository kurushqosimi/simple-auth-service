package verification

import (
	"crypto/rand"
	"fmt"
)

func GenerateOTP() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("cannot generate random bytes: %w", err)
	}

	num := (uint32(b[0]) << 16) | (uint32(b[1]) << 8) | uint32(b[2])
	codeNum := num % 1_000_000

	// Форматируем как 6-значное число, дополняя нулями при необходимости (например "001234").
	codeStr := fmt.Sprintf("%06d", codeNum)

	return codeStr, nil
}
