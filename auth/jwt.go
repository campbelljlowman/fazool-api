package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

// TODO: Get this from env and make it secure!
var secretKey = []byte("EavanRocks!")

func GenerateJWT(id int, authLevel int) (string, error){
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(10 * time.Hour).Unix()
	claims["iss"] = "fazool-api"
	claims["auth"] = authLevel
	claims["user"] = id


	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(bearerToken string) (int, int, error) {
	var tokenString string
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	} else {
		return 0, 0, errors.New("No token passed")
	}


	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return 0, 0, fmt.Errorf("Error parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, 0, fmt.Errorf("JWT Token not valid!")
	}

	id, err := strconv.ParseInt(fmt.Sprintf("%.0f", claims["user"]), 10, 32)
	auth, err := strconv.ParseInt(fmt.Sprintf("%.0f", claims["auth"]), 10, 32)

	return int(id) , int(auth), nil
}

//TODO: Figure out how to refresh tokens