package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/big"
	"net/mail"
	"sync"

	"golang.org/x/exp/slog"
)

func HashHelper(s string) string {
	passwordHashByteArray := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(passwordHashByteArray[:])
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
    return err == nil
}

func GenerateSessionID() (int, error) {
	// Want random 6 digit int, so generate number between 0 and 899999 and then add 100000
	bigNumber, err := rand.Int(rand.Reader, big.NewInt(899999))
	if err != nil {
		return 0, err
	}

    n := bigNumber.Int64()
	number := n + 100000

	return int(number), nil
}

func LogAndReturnError(msg string, err error) error {
	if err != nil {
		slog.Warn(msg, "error", err)
	} else {
		slog.Warn(msg)
	}
	return errors.New(msg)
}

func GetValueFromMutexedMap[T any, V comparable](m map[V]*T, key V, mutex *sync.Mutex) (*T, bool){
	mutex.Lock()
	value, exists := m[key]
	mutex.Unlock()
	return value, exists
}