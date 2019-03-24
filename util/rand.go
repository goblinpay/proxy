package util

import (
	"crypto/rand"
	"encoding/hex"
)

func GetRandomHexString(n int) (string, error) {
	bytes := make([]byte, n / 2) // TODO: implement uneven number of chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
