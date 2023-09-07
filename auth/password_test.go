package auth

import (
	"testing"
)

var passwordFunctionsTests = []struct {
	input string
}{
	{"asdf"},
	{"1234"},
}
func TestPasswordFunctions(t *testing.T) {
	authService := NewAuthService()

	for _, testCase := range(passwordFunctionsTests) {
		passwordHash, err := authService.GenerateBcryptHashForString(testCase.input)

		if err != nil {
			t.Errorf("Error while running HashPassword(%v): %v", testCase.input, err)
		}

		match := authService.CompareBcryptHashAndString(passwordHash, testCase.input)

		//lint:file-ignore S1002 Ignore rule simplifying boolean expression
		if match != true {
			t.Errorf("Password %v doesn't match for hash %v!", testCase.input, passwordHash)
		}
	}
}