package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
)

// TODO: Get this from env and make it secure!
var secretKey = []byte("EavanRocks!")

func GenerateJWT(id int, authLevel int) (string, error){
	println("Creating jwt for ID: ")
	println(id)

	token := jwt.New(jwt.SigningMethodEdDSA)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now()
	claims["exp"] = time.Now().Add(10 * time.Minute)
	claims["iss"] = "fazool-api"
	claims["auth"] = authLevel
	claims["user"] = id

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(token string) ()