package stringx

import (
	"fmt"
	"strconv"
	"time"
)

const minYear = 1900

// ParseReleaseYear parses a year from a string.
// Only the canonical form is accepted: plain digits, no sign, no leading zeros,
// and in reasonable range (1900 - time.Now().Year() + 1).
func ParseReleaseYear(s string) (int16, error) {

	// Convert s to int64. bitSize=16 guarantees s fits in int16
	year, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, err
	}

	// Check range
	if maxYear := time.Now().Year() + 1; year < minYear || int(year) > maxYear {
		return 0, fmt.Errorf("year %d not in range %d-%d", year, minYear, maxYear)
	}

	// Round-trip: rejects "+2024", "002024", etc.
	if strconv.FormatInt(year, 10) != s {
		return 0, fmt.Errorf("year %s is not in canonical form", s)
	}

	return int16(year), nil
}
