package auth

import (
	"os"
	"fmt"
	"time"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/golang-jwt/jwt/v4"
)

var accountTokenDuration time.Duration = 30 // Minutes

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var validAccountClaims = []string{"free", "small-venue", "large-venue"}

func GenerateJWT(userID, accountLevel string) (string, error){
	// Validate inputs
	if (userID == "") {
		return "", fmt.Errorf("User ID is a required field for generating JWT Token!")
	}

	if !slices.Contains(validAccountClaims, accountLevel) {
		return "", fmt.Errorf("Invalid account level for JWT Token! accountLevel: %v", accountLevel)
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(accountTokenDuration * time.Minute).Unix()
	claims["iss"] = "fazool-api"
	claims["user"] = userID
	claims["account-level"] = accountLevel

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(bearerToken string) (string, string, error) {
	var tokenString string
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	} else {
		return "", "", fmt.Errorf("No JWT token passed, token value: %v", bearerToken)
	}


	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("JWT Token not valid! Token: %v", token.Raw)
	}

	id := fmt.Sprintf("%v", claims["user"])
	accountLevel := fmt.Sprintf("%v", claims["account-level"])

	return id, accountLevel, nil
}

//TODO: Figure out how to refresh tokens