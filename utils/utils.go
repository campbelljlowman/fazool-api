package utils

import (
	"net/mail"
	"crypto/sha256"
	"encoding/base64"
)

func HashHelper(s string) string {
	passwordHashByteArray := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(passwordHashByteArray[:])
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
    return err == nil
}