package cep

import "regexp"

var zipcodePattern = regexp.MustCompile(`^\d{8}$`)

func IsValid(zipcode string) bool {
	return zipcodePattern.MatchString(zipcode)
}
