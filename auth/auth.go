package auth

import (
	"os"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"golang.org/x/exp/slog"
)

type AuthService interface {
	GenerateJWTAccessTokenForAccount(accountID int) (string, error)
	GetAccountIDFromJWT(tokenString string) (int, error)

	GetRefreshToken(c *gin.Context)
	RefreshToken(c *gin.Context)

	GenerateBcryptHashForString(inputString string) (string, error)
	CompareBcryptHashAndString(hash, testString string) bool

	CreateAndSendPasswordChangeRequest(email string, accountID int) error
	ValidatePasswordChangeRequest(passwordChangeRequestID string) bool
}

type AuthServiceImpl struct {
	authGorm			*gorm.DB
	smtpServerDomain	string
	smtpServerTLSPort		string
	sourceEmailAddress 	string
	sourceEmailPassword string
	frontendDomain		string
}

func NewAuthService() *AuthServiceImpl{
	smtpServerDomain := os.Getenv("SMTP_SERVER_DOMAIN")
	smtpServerTLSPort := os.Getenv("SMTP_SERVER_TLS_PORT")
	sourceEmailAddress := os.Getenv("SOURCE_EMAIL_ADDRESS")
	sourceEmailPassword := os.Getenv("SOURCE_EMAIL_PASSWORD")
	frontendDomain := os.Getenv("FRONTEND_DOMAIN")

	if (smtpServerDomain == "") || (smtpServerTLSPort == "") || (sourceEmailAddress == "") || (sourceEmailPassword == "") || (frontendDomain == "") {
		slog.Warn("At least one environment variable needed for auth service is empty", 
        "smtpServerDomain", smtpServerDomain, "smtpServerPort", smtpServerTLSPort, 
        "sourceEmailAddress", sourceEmailAddress, "sourceEmailPassword", sourceEmailPassword,
		"frontendDomain", frontendDomain)
		os.Exit(1)
    }

	postgresURL := os.Getenv("POSTGRES_URL")
	slog.Debug("Databse URL", "url", postgresURL)

    gormDB, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{})
	if err != nil {
		slog.Error("Unable to connect to database", err)
		os.Exit(1)
	}

	gormDB.AutoMigrate(&passwordChangeRequest{})

	authServiceImpl := &AuthServiceImpl{
		authGorm: gormDB,
		smtpServerDomain: 		smtpServerDomain,
		smtpServerTLSPort: 		smtpServerTLSPort,
		sourceEmailAddress: 	sourceEmailAddress,
		sourceEmailPassword: 	sourceEmailPassword,
		frontendDomain: 		frontendDomain,
	}

	return authServiceImpl
}