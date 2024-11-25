package common

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Define the regex patterns as constants
const (
	phoneRegex   = `^(?:\+254|254|0)?((?:7[0-9]{8})|(?:1[01][0-9]{7}))$`
	addressRegex = `^0x[a-fA-F0-9]{40}$`
	aliasRegex   = `^[a-zA-Z0-9]+$`
)

// IsValidPhoneNumber checks if the given number is a valid phone number
func IsValidPhoneNumber(phonenumber string) bool {
	match, _ := regexp.MatchString(phoneRegex, phonenumber)
	return match
}

// IsValidAddress checks if the given address is a valid Ethereum address
func IsValidAddress(address string) bool {
	match, _ := regexp.MatchString(addressRegex, address)
	return match
}

// IsValidAlias checks if the alias is a valid alias format
func IsValidAlias(alias string) bool {
	match, _ := regexp.MatchString(aliasRegex, alias)
	return match
}

// CheckRecipient validates the recipient format based on the criteria
func CheckRecipient(recipient string) (string, error) {
	if IsValidPhoneNumber(recipient) {
		return "phone number", nil
	}

	if IsValidAddress(recipient) {
		return "address", nil
	}

	if IsValidAlias(recipient) {
		return "alias", nil
	}

	return "", fmt.Errorf("invalid recipient: must be a phone number, address or alias")
}

// FormatPhoneNumber formats a Kenyan phone number to "+254xxxxxxxx".
func FormatPhoneNumber(phone string) (string, error) {
	if !IsValidPhoneNumber(phone) {
		return "", errors.New("invalid phone number")
	}

	// Remove any leading "+" and spaces
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.ReplaceAll(phone, " ", "")

	// Replace leading "0" with "254" if present
	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	// Add "+" if not already present
	if !strings.HasPrefix(phone, "254") {
		return "", errors.New("unexpected format")
	}

	return "+" + phone, nil
}
