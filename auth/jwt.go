package auth

import (
	"os"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const accountTokenDuration time.Duration = 30 // Minutes

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWT(userID string) (string, error){
	// Validate inputs
	if (userID == "") {
		return "", fmt.Errorf("User ID is a required field for generating JWT Token!")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(accountTokenDuration * time.Minute).Unix()
	claims["iss"] = "fazool-api"
	claims["user"] = userID

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("JWT Token not valid! Token: %v", token.Raw)
	}

	id := fmt.Sprintf("%v", claims["user"])

	return id, nil
}

//TODO: Figure out how to refresh tokens