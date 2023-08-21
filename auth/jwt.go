package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slog"
)

const accountTokenDurationMinutes time.Duration = 30

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWTForAccount(accountID int) (string, error){
	if (accountID == 0) {
		return "", fmt.Errorf("account ID is a required field for generating JWT Token")
	}

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	jwtClaims := jwtToken.Claims.(jwt.MapClaims)
	jwtClaims["iat"] = time.Now().Unix()
	jwtClaims["exp"] = time.Now().Add(accountTokenDurationMinutes * time.Minute).Unix()
	jwtClaims["iss"] = "fazool-api"
	jwtClaims["accountID"] = accountID

	jwtTokenString, err := jwtToken.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return jwtTokenString, nil
}

func GetAccountIDFromJWT(tokenString string) (int, error) {
	jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return 0, fmt.Errorf("jwt is expired")
	}
	if err != nil {
		return 0, err
	}

	jwtClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return 0, fmt.Errorf("JWT Token not valid! Token: %v", jwtToken.Raw)
	}

	accountID, err := strconv.Atoi(fmt.Sprintf("%v", jwtClaims["accountID"]))

	if err != nil {
		slog.Warn("Error converting account ID passed on JWT token to int")
	}

	return accountID, nil
}

//TODO: Figure out how to refresh tokens