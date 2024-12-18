package utils

import (
	"strconv"
	"time"
)

// CalculateAge calculates the age based on a given birthdate and the current date in the format dd/mm/yy
// It adjusts for cases where the current date is before the birthday in the current year.
func CalculateAge(birthdate, today time.Time) int {
	today = today.In(birthdate.Location())
	ty, tm, td := today.Date()
	today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := birthdate.Date()
	birthdate = time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birthdate) {
		return 0
	}
	age := ty - by
	anniversary := birthdate.AddDate(age, 0, 0)
	if anniversary.After(today) {
		age--
	}
	return age
}

// CalculateAgeWithYOB calculates the age based on the given year of birth (YOB).
// It subtracts the YOB from the current year to determine the age.
//
// Parameters:
//
//	yob: The year of birth as an integer.
//
// Returns:
//
//	The calculated age as an integer.
func CalculateAgeWithYOB(yob int) int {
	currentYear := time.Now().Year()
	return currentYear - yob
}


//IsValidYob checks if the provided yob can be considered valid 
func IsValidYOb(yob string) bool {
	currentYear := time.Now().Year()
	yearOfBirth, err := strconv.ParseInt(yob, 10, 64)
	if err != nil {
		return false
	}
	if yearOfBirth >= 1900 && int(yearOfBirth) <= currentYear {
		return true
	} else {
		return false
	}

}
