package utils

import (
	"errors"
	"testing"
)

var hashHelperTests = []struct {
	input string
	expectedOutput string
}{
	{"asdf", "8OTC92xYkW7CWPJGhRvqCR0U1CR6L8PhhpRGGxgW4Ts="},
	{"", "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU="},
}
func TestHashHelper(t *testing.T) {
	for _, testCase := range(hashHelperTests) {
		resultString := HashHelper(testCase.input)

		if resultString != testCase.expectedOutput {
			t.Errorf("HashHelper(%v) failed! Wanted: %v, got: %v", testCase.input, testCase.expectedOutput, resultString)
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