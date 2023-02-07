package auth

import (
	"os"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const accountTokenDurationMinutes time.Duration = 30

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWTForAccount(accountID string) (string, error){
	// Validate inputs
	if (accountID == "") {
		return "", fmt.Errorf("Account ID is a required field for generating JWT Token!")
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

func GetAccountIDFromJWT(tokenString string) (string, error) {
	jwtToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})
	if err != nil {
		return "", err
	}

	jwtClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return "", fmt.Errorf("JWT Token not valid! Token: %v", jwtToken.Raw)
	}

	accountID := fmt.Sprintf("%v", jwtClaims["accountID"])

	return accountID, nil
}

//TODO: Figure out how to refresh tokens