package common

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Define the regex pattern as a constant
	pinPattern = `^\d{4}$`

	//Allowed incorrect  PIN attempts
	AllowedPINAttempts = uint8(3)
	
)

// checks whether the given input is a 4 digit number
func IsValidPIN(pin string) bool {
	match, _ := regexp.MatchString(pinPattern, pin)
	return match
}

// HashPIN uses bcrypt with 8 salt rounds to hash the PIN
func HashPIN(pin string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), 8)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPIN compareS the hashed PIN with the plaintext PIN
func VerifyPIN(hashedPIN, pin string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPIN), []byte(pin))
	return err == nil
}
