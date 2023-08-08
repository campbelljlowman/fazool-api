package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/mail"

	"golang.org/x/exp/slog"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	passwordBytes := []byte(password)

	passwordHash, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.MinCost)

	return string(passwordHash), err
}

func CompareHashAndPassword(passwordHash, password string) bool {
	passwordHashBytes := []byte(passwordHash)
	passwordBytes := []byte(password)

	err := bcrypt.CompareHashAndPassword(passwordHashBytes, passwordBytes)
	return err == nil
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
		slog.Error(msg, err)
	} else {
		slog.Warn(msg)
	}
	return errors.New(msg)
}