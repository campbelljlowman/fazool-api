package auth

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/campbelljlowman/fazool-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slog"
)

const accountAccessTokenDurationMinutes time.Duration = 30
const accountRefreshTokenDurationHours time.Duration = 168 // 24 * 7
const refreshTokenCookieName string = "fazool_refresh_token"
const refreshTokenCookieMaxAgeSeconds int = 608400 // 60 seconds * 60 minutes * 24 hours * 7 days + 3600 seconds 

var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateJWTAccessTokenForAccount(accountID int) (string, error){
	accessToken, err := generateJWTForAccount(accountID, accountAccessTokenDurationMinutes * time.Minute) 
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func generateJWTForAccount(accountID int, duractionValid time.Duration) (string, error){
	if (accountID == 0) {
		return "", fmt.Errorf("account ID is a required field for generating JWT Token")
	}

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	jwtClaims := jwtToken.Claims.(jwt.MapClaims)
	jwtClaims["iat"] = time.Now().Unix()
	jwtClaims["exp"] = time.Now().Add(duractionValid).Unix()
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

func GetRefreshToken(c *gin.Context) {
	accoundID, _ := c.Request.Context().Value("accountID").(int)
	if accoundID == 0 {
		c.AbortWithError(403, fmt.Errorf("account ID couldn't be parsed from request: %v", accoundID))
		return
	}

	refreshToken, err := generateJWTForAccount(accoundID, accountRefreshTokenDurationHours * time.Hour)
	if err != nil {
		utils.LogAndReturnError("error generating refresh token", err)
		c.Abort()
	}

	frontendDomain := os.Getenv("FRONTEND_DOMAIN")
	c.SetCookie(refreshTokenCookieName, refreshToken, refreshTokenCookieMaxAgeSeconds, "/", frontendDomain, true, true)
}

func RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		utils.LogAndReturnError("error getting refresh token cookie", err)
		c.Abort()
		return
	}

	accountID, err := GetAccountIDFromJWT(refreshToken)
	if err != nil {
		utils.LogAndReturnError("error getting accound ID from refresh token", err)
		c.Abort()
		return
	}

	slog.Debug("refreshing access token", "account", accountID)

	accessToken, err := generateJWTForAccount(accountID, accountAccessTokenDurationMinutes * time.Minute)
	if err != nil {
		utils.LogAndReturnError("error generating new access token", err)
		c.Abort()
		return
	}

	c.String(200, accessToken)
}