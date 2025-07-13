package sitemaps

import (
	"unicode"
)

func isDigitsOnly(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validateDate(year, month string) bool {
	return len(year) == 4 && len(month) == 2 && isDigitsOnly(year+month)
}
