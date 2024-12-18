package utils

func ConstructName(firstName, familyName, defaultValue string) string {
	name := defaultValue
	if familyName != defaultValue {
		if firstName != defaultValue {
			name = firstName + " " + familyName
		} else {
			name = familyName
		}
	} else {
		if firstName != defaultValue {
			name = firstName
		}
	}
	return name
}
