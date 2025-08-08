package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateLaravelAppKey() (string, error) {
	buf := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("no se pudo generar APP_KEY: %w", err)
	}
	return "base64:" + base64.StdEncoding.EncodeToString(buf), nil
}
