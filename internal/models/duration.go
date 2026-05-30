package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Custom string type used to convert string duration to desirable format
type ISO8601Duration string

// Valid ISO time format
var validISO8601 = regexp.MustCompile(`(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?`)

// Compile an ISO-8601 string to time components
func (d ISO8601Duration) compile() (map[string]int, error) {
	// Check for PT prefix
	if !strings.HasPrefix(string(d), "PT") {
		return nil, fmt.Errorf("invalid duration format: %s", d)
	}

	// Remove the PT prefix
	duration := strings.TrimPrefix(string(d), "PT")

	// Find the substrings (hours, minutes, seconds)
	matches := validISO8601.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid duration format: %s", duration)
	}

	// Check for the matched regex groups
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	sec, _ := strconv.ParseFloat(matches[3], 64)
	seconds := int(sec)

	return map[string]int{
		"h": hours,
		"m": minutes,
		"s": seconds,
	}, nil
}

// Get human readbale video duration
func (d ISO8601Duration) Human() (string, error) {
	t, err := d.compile()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%02d:%02d:%02d", t["h"], t["m"], t["s"]), nil
}

// Get video duration in seconds
func (d ISO8601Duration) Seconds() (time.Duration, error) {
	t, err := d.compile()
	if err != nil {
		return 0, err
	}

	seconds := t["h"]*60*60 + t["m"]*60 + t["s"]
	return time.Duration(seconds) * time.Second, nil
}
