package common

import (
	"encoding/hex"
	"strings"
)

func NormalizeHex(s string) (string, error) {
	if len(s) >= 2 {
		if s[:2] == "0x" {
			s = s[2:]
		}
	}
	r, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(r), nil
}

func IsSameHex(left string, right string) bool {
	bl, err := NormalizeHex(left)
	if err != nil {
		return false
	}
	br, err := NormalizeHex(left)
	if err != nil {
		return false
	}
	return strings.Compare(bl, br) == 0
}
