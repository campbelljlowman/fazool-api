package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func HashHelper(s string) string {
	passwordHashByteArray := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(passwordHashByteArray[:])
}