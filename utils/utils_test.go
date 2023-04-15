package utils

import (
	"errors"
	"testing"
)

var passwordFunctionsTests = []struct {
	input string
}{
	{"asdf"},
	{"1234"},
}
func TestPasswordFunctions(t *testing.T) {
	for _, testCase := range(passwordFunctionsTests) {
		passwordHash, err := HashPassword(testCase.input)

		if err != nil {
			t.Errorf("Error while running HashPassword(%v): %v", testCase.input, err)
		}

		match := CompareHashAndPassword(passwordHash, testCase.input)

		if match != true {
			t.Errorf("Password %v doesn't match for hash %v!", testCase.input, passwordHash)
		}
	}
}

var validateEmailTests = []struct {
	input string
	expectedOutput bool
}{
	{"a@a.a", true},
	{"campbell@gmail.com", true},
	{"a", false},
}
func TestValidateEmail(t *testing.T){
	for _, testCase := range(validateEmailTests) {
		resultBool := ValidateEmail(testCase.input)

		if resultBool != testCase.expectedOutput {
			t.Errorf("ValidateEmail(%v) failed! Wanted: %v, got: %v", testCase.input, testCase.expectedOutput, resultBool)
		}
	}
}

func TestGenerateSessionID(t *testing.T){
	sessionID, err := GenerateSessionID()

	if sessionID < 100000 || sessionID > 999999 || err != nil {
		t.Errorf("GenerateSessionID() failed! Value: %v, Error: %v", sessionID, err)
	}
}

var logAndReturnErrorTests = []struct {
	inputMessage string
	inputError error
	expectedOutput error
}{
	{"Test error message", nil, errors.New("Test error message")},
	{"", nil, errors.New("")},
	{"Test error message", errors.New("Error message to log"),  errors.New("Test error message")},
}
func TestLogAndReturnError(t *testing.T) {
	for _, testCase := range(logAndReturnErrorTests) {
		resultError := LogAndReturnError(testCase.inputMessage, testCase.inputError)

		if resultError.Error() != testCase.expectedOutput.Error() {
			t.Errorf("LogAndReturnError(%v, %v) failed! Wanted: %v, got: %v", testCase.inputMessage, testCase.inputError, testCase.expectedOutput, resultError)
		}
	}
}