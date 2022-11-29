package auth

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)


var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWT(sessionID int, userID, accountLevel, voterLevel string) (string, error){
	// TODO: Add login in here for the different claims possibilties
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(10 * time.Hour).Unix()
	claims["iss"] = "fazool-api"
	claims["account-level"] = accountLevel
	claims[strconv.Itoa(sessionID)] = voterLevel
	claims["user"] = userID


	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(bearerToken string) (string, int, error) {
	var tokenString string
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	} else {
		return "", 0, fmt.Errorf("No JWT token passed, token value: %v", bearerToken)
	}


	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return "", 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", 0, fmt.Errorf("JWT Token not valid! Token: %v", token.Raw)
	}

	id := fmt.Sprintf("%v", claims["user"])
	auth, err := strconv.ParseInt(fmt.Sprintf("%.0f", claims["auth"]), 10, 32)

	return id, int(auth), nil
}

//TODO: Figure out how to refresh tokens