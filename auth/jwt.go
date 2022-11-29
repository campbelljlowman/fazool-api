package auth

import (
	"os"
	"fmt"
	"time"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/golang-jwt/jwt/v4"
)

// TODO: Figure out how long these should be
var accountTokenDuration time.Duration = 8 // Hours
var voterTokenDuration time.Duration = 10 // Minutes
var privilegedVoterTokenDuration time.Duration = 30 // Minutes

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var validVoterClaims = []string{"voter", "privileged-voter", "admin"}
var validAccountClaims = []string{"free", "small-venue", "large-venue"}

func GenerateJWT(sessionID int, userID, accountLevel, voterLevel string) (string, error){
	// Validate inputs
	voterClaims := VoterClaims{
		sessionID: sessionID,
		level: voterLevel,
	}

	if userID == "" {
		return "", fmt.Errorf("User ID is a required field for generating JWT Token")
	}

	if (accountLevel == "") && voterClaims.empty() {
		return "", fmt.Errorf("Account claims or voter claims are required for generating JWT Token")
	}


	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["iss"] = "fazool-api"
	claims["user"] = userID

	if (accountLevel != "") && slices.Contains(validAccountClaims, accountLevel) {
		claims["account-level"] = accountLevel
		claims["exp"] = time.Now().Add(accountTokenDuration * time.Hour).Unix()
	}

	if !voterClaims.empty() {
		if voterClaims.valid() {
			claims["voter-level"] = voterClaims

			// If voter token is admin, keep account token expiration time
			if voterClaims.level == "privileged-voter" {
				claims["exp"] = time.Now().Add(privilegedVoterTokenDuration * time.Minute).Unix()
			}else if voterClaims.level == "voter" {
				claims["exp"] = time.Now().Add(voterTokenDuration * time.Minute).Unix()
			}
			
		} else {
			return "", fmt.Errorf(`Must pass sessionID and voter level for JWT voter claims. 
			sessionID: %v, voter level: %v`, voterClaims.sessionID, voterClaims.level)
		}
	}


	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(bearerToken string) (string, string, *VoterClaims, error) {
	var tokenString string
	if len(strings.Split(bearerToken, " ")) == 2 {
		tokenString = strings.Split(bearerToken, " ")[1]
	} else {
		return "", "", nil, fmt.Errorf("No JWT token passed, token value: %v", bearerToken)
	}


	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return "", "", nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", nil, fmt.Errorf("JWT Token not valid! Token: %v", token.Raw)
	}

	id := fmt.Sprintf("%v", claims["user"])
	accountLevel := fmt.Sprintf("%v", claims["account-level"])
	voterClaims := claims["voter-level"].(VoterClaims)

	return id, accountLevel, &voterClaims, nil
}

//TODO: Figure out how to refresh tokens

type VoterClaims struct {
	sessionID int
	level string
}

func (v VoterClaims) empty() bool {
	return (v.sessionID == 0) && (v.level == "")
}

func (v VoterClaims) valid() bool {
	return ((v.sessionID == 0) != (v.level == "")) || !(slices.Contains(validVoterClaims, v.level))
}
