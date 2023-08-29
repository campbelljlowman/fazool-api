package auth

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

func (a *AuthServiceImpl) GenerateBcryptHashForString(inputString string) (string, error) {
	inputStringBytes := []byte(inputString)

	bcryptHash, err := bcrypt.GenerateFromPassword(inputStringBytes, bcrypt.MinCost)

	return string(bcryptHash), err
}

func (a *AuthServiceImpl) CompareBcryptHashAndString(hash, testString string) bool {
	hashBytes := []byte(hash)
	testStringBytes := []byte(testString)

	err := bcrypt.CompareHashAndPassword(hashBytes, testStringBytes)
	return err == nil
}

type passwordChangeRequest struct {
	gorm.Model
	passwordChangeRequestIDHash string
	accountID 					int
	timeCreated					time.Time
}

func (a *AuthServiceImpl) CreateAndSendPasswordChangeRequest(email string, accountID int) error {
	passwordChangeRequestID := uuid.New().String()
	passwordChangeRequestIDHash, err := a.GenerateBcryptHashForString(passwordChangeRequestID)
	if err != nil {
		return err
	}

	slog.Debug("Password change request id hash", "hash", passwordChangeRequestIDHash)
	passwordChangeRequest := passwordChangeRequest{
		passwordChangeRequestIDHash: 	passwordChangeRequestIDHash,
		accountID: 						accountID,
		timeCreated: 					time.Now(),
	}
	a.authGorm.Create(&passwordChangeRequest)

	err = a.sendPasswordChangeRequestEmail(email, passwordChangeRequestID)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthServiceImpl) sendPasswordChangeRequestEmail(destinationEmail, passwordChangeRequestID string) error {
	subject := "Fazool Password Reset"
	message := fmt.Sprintf("A password reset has been requested for your Fazool account. " +
	"To reset your password, navigate to the folling link: %v/change-password?passwordChangeRequestID=%v", a.frontendDomain, passwordChangeRequestID)

	emailBody := "To: " + destinationEmail + "\r\n" +
	"Subject: " + subject + "\r\n" +
	"\r\n" +
	message + "\r\n"
	slog.Debug("Recovery email", "email", emailBody)

	auth := smtp.PlainAuth("", a.sourceEmailAddress, a.sourceEmailPassword, a.smtpServerDomain)

	err := smtp.SendMail(fmt.Sprintf("%s:%s", a.smtpServerDomain, a.smtpServerTLSPort), auth, a.sourceEmailAddress, []string{destinationEmail}, []byte(emailBody))
	
	slog.Debug("Send recovery email result", "result", err)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthServiceImpl) ValidatePasswordChangeRequest(passwordChangeRequestID string) bool {
	return false
}